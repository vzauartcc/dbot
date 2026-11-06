package commands

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

const pirepUnknown = "UNKN"

var (
	errPirepType             = errors.New("invalid PIREP type")
	errPirepLocation         = errors.New("invalid /OV location")
	errPirepTime             = errors.New("invalid /TM time")
	errPirepFlightLevel      = errors.New("invalid /FL flight level")
	errPirepAircraftType     = errors.New("invalid /TP aircraft type")
	errPirepSky              = errors.New("invalid /SK sky conditions")
	errPirepTemperature      = errors.New("invalid /TA temperature")
	errPirepWind             = errors.New("invalid /WV wind")
	errPirepIcingTemperature = errors.New(
		"air temperature (/TA) is required when reporting icing (/IC)",
	)
)

var pirepSkyCovers = map[string]struct{}{
	"SKC": {},
	"CLR": {},
	"FEW": {},
	"SCT": {},
	"BKN": {},
	"OVC": {},
	"OVX": {},
}

var Pirep = &discordgo.ApplicationCommand{
	Name:        "pirep",
	Description: "Submit a Pirep",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "ov",
			Description: "Observation Location (fix, coordinates, or bearing/distance), use 3-letter NAVAID idents only",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "type",
			Description: "PIREP type",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "UA - Routine", Value: "UA"},
				{Name: "UUA - Urgent", Value: "UUA"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "tm",
			Description: "Observation time in UTC (HHMM)",
			Required:    true,
			MinLength:   new(4),
			MaxLength:   4,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "fl",
			Description: "Flight level in hundreds of feet (e.g. 095) or UNKN",
			Required:    true,
			MinLength:   new(3),
			MaxLength:   4,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "tp",
			Description: "Aircraft type ICAO code (e.g. B737), or UNKN",
			Required:    true,
			MaxLength:   4,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "sk",
			Description: "Sky conditions (e.g. SCT030 BKN 070)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "wx",
			Description: "Weather and visibility (e.g. 10SM RA), intensity is not reported",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "ta",
			Description: "Air temperature in Celsius (e.g. -03, 13)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "wv",
			Description: "Wind direction and speed (e.g. 270011)",
			MaxLength:   6,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "tb",
			Description: "Turbulence type, frequency and intensity, using standard contractions",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "ic",
			Description: "Icing type and intensity, using standard intensity and contractions",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "rm",
			Description: "Additional remarks",
		},
	},
}

func PirepHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "pirep", true) {
		return
	}

	// debug: change to real channel later
	if i.ChannelID != "1423816312299196426" {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "This command may not be used in this channel.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending invalid channel response for /pirep for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	opts := make(map[string]string)
	for _, opt := range i.ApplicationCommandData().Options {
		opts[opt.Name] = opt.StringValue()
	}

	report, err := buildPirep(opts)
	if err != nil {
		_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: err.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending validation response for /pirep for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
		Content: report,
	})
	if err != nil {
		log.Printf(
			"Error sending response for /pirep for %s: %v\n",
			helpers.GetMemberName(i.Member),
			err,
		)
	}
}

func buildPirep(opts map[string]string) (string, error) {
	if strings.TrimSpace(opts["ic"]) != "" && strings.TrimSpace(opts["ta"]) == "" {
		return "", errPirepIcingTemperature
	}

	reportType, err := parsePirepType(opts["type"])
	if err != nil {
		return "", err
	}

	location, err := parsePirepLocation(opts["ov"])
	if err != nil {
		return "", err
	}

	hour, minute, err := parsePirepTime(opts["tm"])
	if err != nil {
		return "", err
	}

	altitude, unknownAltitude, err := parsePirepFlightLevel(opts["fl"])
	if err != nil {
		return "", err
	}

	aircraftType, err := parsePirepAircraftType(opts["tp"])
	if err != nil {
		return "", err
	}

	var report strings.Builder

	_, _ = fmt.Fprintf(
		&report,
		"**PIREP**\n`%s /OV %s/TM %02d%02d",
		reportType,
		formatPirepLocation(location),
		hour,
		minute,
	)

	if unknownAltitude {
		_, _ = fmt.Fprintf(&report, "/FL%s", pirepUnknown)
	} else {
		_, _ = fmt.Fprintf(&report, "/FL%03d", altitude/100)
	}

	_, _ = fmt.Fprintf(&report, "/TP %s", aircraftType)

	if value := strings.TrimSpace(opts["sk"]); value != "" {
		layers, err := parsePirepSky(value)
		if err != nil {
			return "", err
		}

		_, _ = fmt.Fprintf(&report, "/SK %s", formatPirepSky(layers))
	}

	writePirepField(&report, "WX", opts["wx"])

	if value := strings.TrimSpace(opts["ta"]); value != "" {
		temperature, err := parsePirepTemperature(value)
		if err != nil {
			return "", err
		}

		if temperature < 0 {
			_, _ = fmt.Fprintf(&report, "/TA -%02d", -temperature)
		} else {
			_, _ = fmt.Fprintf(&report, "/TA %02d", temperature)
		}
	}

	if value := strings.TrimSpace(opts["wv"]); value != "" {
		direction, speed, err := parsePirepWind(value)
		if err != nil {
			return "", err
		}

		_, _ = fmt.Fprintf(&report, "/WV %03d%03d", direction, speed)
	}

	writePirepField(&report, "TB", opts["tb"])
	writePirepField(&report, "IC", opts["ic"])
	writePirepField(&report, "RM", opts["rm"])

	_, _ = fmt.Fprint(&report, "`")

	return report.String(), nil
}

// writePirepField appends an upper-cased free text group when a value is reported.
func writePirepField(report *strings.Builder, label, value string) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return
	}

	_, _ = fmt.Fprintf(report, "/%s %s", label, strings.ToUpper(strings.Join(fields, " ")))
}

// parsePirepType validates the report type and returns the encoded form.
func parsePirepType(value string) (string, error) {
	reportType := strings.ToUpper(strings.TrimSpace(value))
	if reportType != "UA" && reportType != "UUA" {
		return "", fmt.Errorf("%w: %q must be UA or UUA", errPirepType, value)
	}

	return reportType, nil
}

// pirepFix is one point in a parsed /OV location.
type pirepFix struct {
	Ident              string
	Bearing            int
	Distance           int
	HasBearingDistance bool
}

// parsePirepLocation parses an /OV location into its fixes.
func parsePirepLocation(ov string) ([]pirepFix, error) {
	fields := strings.Fields(ov)
	if len(fields) == 0 {
		return nil, fmt.Errorf("%w: a fix is required", errPirepLocation)
	}

	fixes := make([]pirepFix, 0, len(fields))

	for _, field := range fields {
		for part := range strings.SplitSeq(field, "-") {
			part = strings.ToUpper(part)

			if isPirepDigits(part, 6) {
				if len(fixes) == 0 {
					return nil, fmt.Errorf("%w: %q has no preceding fix", errPirepLocation, part)
				}

				fix := &fixes[len(fixes)-1]
				if fix.HasBearingDistance {
					return nil, fmt.Errorf(
						"%w: %q repeats a bearing and distance",
						errPirepLocation,
						part,
					)
				}

				bearing := pirepNumber(part[:3])
				if bearing > 360 {
					return nil, fmt.Errorf(
						"%w: bearing %03d is out of range",
						errPirepLocation,
						bearing,
					)
				}

				fix.Bearing = bearing
				fix.Distance = pirepNumber(part[3:])
				fix.HasBearingDistance = true

				continue
			}

			fix, err := parsePirepFix(part)
			if err != nil {
				return nil, err
			}

			fixes = append(fixes, fix)
		}
	}

	return fixes, nil
}

// parsePirepFix parses a single /OV fix with an optional attached bearing and distance.
func parsePirepFix(part string) (pirepFix, error) {
	if len(part) < 3 || !isPirepIdent(part[:3]) {
		return pirepFix{}, fmt.Errorf(
			"%w: %q must start with a 3-character NAVAID ident",
			errPirepLocation,
			part,
		)
	}

	fix := pirepFix{Ident: part[:3]}

	rest := part[3:]
	if rest == "" {
		return fix, nil
	}

	if !isPirepDigits(rest, 6) {
		return pirepFix{}, fmt.Errorf(
			"%w: %q must be a 6-digit bearing and distance",
			errPirepLocation,
			rest,
		)
	}

	bearing := pirepNumber(rest[:3])
	if bearing > 360 {
		return pirepFix{}, fmt.Errorf("%w: bearing %03d is out of range", errPirepLocation, bearing)
	}

	fix.Bearing = bearing
	fix.Distance = pirepNumber(rest[3:])
	fix.HasBearingDistance = true

	return fix, nil
}

// formatPirepLocation renders parsed fixes in the /OV encoding.
func formatPirepLocation(fixes []pirepFix) string {
	var report strings.Builder

	for idx, fix := range fixes {
		if idx > 0 {
			_, _ = fmt.Fprint(&report, "-")
		}

		_, _ = fmt.Fprint(&report, fix.Ident)

		if fix.HasBearingDistance {
			_, _ = fmt.Fprintf(&report, " %03d%03d", fix.Bearing, fix.Distance)
		}
	}

	return report.String()
}

// parsePirepTime validates a /TM time and returns the hour and minute.
func parsePirepTime(timeValue string) (int, int, error) {
	if !isPirepDigits(timeValue, 4) {
		return 0, 0, fmt.Errorf("%w: %q must be 4 digits (HHMM)", errPirepTime, timeValue)
	}

	hour := pirepNumber(timeValue[:2])
	minute := pirepNumber(timeValue[2:])

	if hour > 23 || minute > 59 {
		return 0, 0, fmt.Errorf("%w: %q is not a valid time", errPirepTime, timeValue)
	}

	return hour, minute, nil
}

// parsePirepFlightLevel returns the reported altitude in feet. Unknown is true when the
// flight level is reported as UNKN.
func parsePirepFlightLevel(flightLevel string) (int, bool, error) {
	flightLevel = strings.ToUpper(strings.TrimSpace(flightLevel))

	if flightLevel == pirepUnknown {
		return 0, true, nil
	}

	if !isPirepDigits(flightLevel, 3) {
		return 0, false, fmt.Errorf(
			"%w: %q must be 3 digits (hundreds of feet) or %s",
			errPirepFlightLevel,
			flightLevel,
			pirepUnknown,
		)
	}

	return pirepNumber(flightLevel) * 100, false, nil
}

// parsePirepAircraftType validates a /TP aircraft type and returns the encoded form.
func parsePirepAircraftType(aircraft string) (string, error) {
	aircraftType := strings.ToUpper(strings.TrimSpace(aircraft))

	if aircraftType == pirepUnknown {
		return aircraftType, nil
	}

	if len(aircraftType) == 0 || len(aircraftType) > 4 {
		return "", fmt.Errorf(
			"%w: %q must be at most 4 characters or %s",
			errPirepAircraftType,
			aircraft,
			pirepUnknown,
		)
	}

	for _, char := range aircraftType {
		if !isPirepAlnum(char) {
			return "", fmt.Errorf(
				"%w: %q must contain only letters and digits",
				errPirepAircraftType,
				aircraft,
			)
		}
	}

	return aircraftType, nil
}

// pirepSkyLayer is one cloud layer in a parsed /SK report.
type pirepSkyLayer struct {
	Base  string
	Cover string
	Tops  string
}

// parsePirepSky parses a /SK report into its cloud layers.
func parsePirepSky(sk string) ([]pirepSkyLayer, error) {
	rawLayers := strings.Split(sk, "/")
	layers := make([]pirepSkyLayer, 0, len(rawLayers))

	for _, rawLayer := range rawLayers {
		layer, err := parsePirepSkyLayer(rawLayer)
		if err != nil {
			return nil, err
		}

		layers = append(layers, layer)
	}

	return layers, nil
}

// parsePirepSkyLayer parses a single /SK layer.
func parsePirepSkyLayer(rawLayer string) (pirepSkyLayer, error) {
	fields := strings.Fields(strings.ToUpper(rawLayer))
	if len(fields) < 2 || len(fields) > 3 {
		return pirepSkyLayer{}, fmt.Errorf(
			"%w: %q must report a base height, a cover and optional tops",
			errPirepSky,
			rawLayer,
		)
	}

	base := fields[0]
	if base != pirepUnknown && !isPirepDigits(base, 3) {
		return pirepSkyLayer{}, fmt.Errorf(
			"%w: base height %q must be 3 digits or %s",
			errPirepSky,
			base,
			pirepUnknown,
		)
	}

	if _, ok := pirepSkyCovers[fields[1]]; !ok {
		return pirepSkyLayer{}, fmt.Errorf("%w: unknown cloud cover %q", errPirepSky, fields[1])
	}

	layer := pirepSkyLayer{Base: base, Cover: fields[1]}

	if len(fields) == 3 {
		if !isPirepDigits(fields[2], 3) {
			return pirepSkyLayer{}, fmt.Errorf(
				"%w: cloud tops %q must be 3 digits",
				errPirepSky,
				fields[2],
			)
		}

		layer.Tops = fields[2]
	}

	return layer, nil
}

// formatPirepSky renders parsed cloud layers in the /SK encoding.
func formatPirepSky(layers []pirepSkyLayer) string {
	rendered := make([]string, 0, len(layers))

	for _, layer := range layers {
		if layer.Tops != "" {
			rendered = append(rendered, layer.Base+" "+layer.Cover+" "+layer.Tops)

			continue
		}

		rendered = append(rendered, layer.Base+" "+layer.Cover)
	}

	return strings.Join(rendered, "/")
}

// parsePirepTemperature returns the /TA air temperature in Celsius.
func parsePirepTemperature(temperatureValue string) (int, error) {
	value := strings.TrimSpace(temperatureValue)

	sign := 1
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.TrimPrefix(value, "-")
	}

	if len(value) < 1 || len(value) > 3 || !isPirepDigits(value, len(value)) {
		return 0, fmt.Errorf(
			"%w: %q must be a temperature in Celsius",
			errPirepTemperature,
			temperatureValue,
		)
	}

	temperature := sign * pirepNumber(value)
	if temperature < -99 || temperature > 99 {
		return 0, fmt.Errorf("%w: %q is out of range", errPirepTemperature, temperatureValue)
	}

	return temperature, nil
}

// parsePirepWind validates a /WV wind and returns the direction and speed.
func parsePirepWind(wind string) (int, int, error) {
	if !isPirepDigits(wind, 6) {
		return 0, 0, fmt.Errorf(
			"%w: %q must be 6 digits (direction and speed)",
			errPirepWind,
			wind,
		)
	}

	direction := pirepNumber(wind[:3])
	if direction > 360 {
		return 0, 0, fmt.Errorf("%w: direction %03d is out of range", errPirepWind, direction)
	}

	return direction, pirepNumber(wind[3:]), nil
}

// isPirepDigits reports whether s is exactly length decimal digits.
func isPirepDigits(s string, length int) bool {
	if len(s) != length {
		return false
	}

	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// isPirepIdent reports whether s is a 3-character alphanumeric NAVAID ident.
func isPirepIdent(s string) bool {
	if len(s) != 3 {
		return false
	}

	for _, char := range s {
		if !isPirepAlnum(char) {
			return false
		}
	}

	return true
}

// isPirepAlnum reports whether r is an upper-case letter or a digit.
func isPirepAlnum(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// pirepNumber converts digits already validated by isPirepDigits.
func pirepNumber(s string) int {
	value, _ := strconv.Atoi(s)

	return value
}
