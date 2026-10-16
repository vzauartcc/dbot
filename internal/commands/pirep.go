package commands

import (
	"fmt"
	"log"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var Pirep = &discordgo.ApplicationCommand{
	Name:        "pirep",
	Description: "Submit a Pirep",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "ov",
			Description: "Observation Location (fix, coordinates, or bearing/distance)",
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
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "fl",
			Description: "Flight level in hundreds of feet (e.g. 095)",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "tp",
			Description: "Aircraft type ICAO code (e.g. B737)",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "sk",
			Description: "Sky conditions (e.g. SCT030 BKN 070)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "wx",
			Description: "Weather and visibility (e.g. 10SM RA)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "ta",
			Description: "Air temperature in Celsius (e.g. M03)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "wv",
			Description: "Wind direction and speed (e.g. 270011KT)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "tb",
			Description: "Turbulence type, frequency and intensity",
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "ic",
			Description: "Icing type and intensity",
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

	if opts["ic"] != "" && opts["tm"] == "" {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Air temperature (/TA) is required when reporting icing (/IC)",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending valiation response for /pirep for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	fields := []struct {
		key   string
		label string
	}{
		{"ov", "OV"},
		{"tm", "TM"},
		{"fl", "FL"},
		{"tp", "TP"},
		{"sk", "SK"},
		{"wx", "WX"},
		{"ta", "TA"},
		{"wv", "WV"},
		{"tb", "TB"},
		{"ic", "IC"},
		{"rm", "RM"},
	}

	loc := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return -1
		}

		return r
	}, opts["ov"])

	var report strings.Builder
	fmt.Fprintf(&report, "**PIREP**\n`%s %s ", loc, opts["type"])

	for _, field := range fields {
		if value := opts[field.key]; value != "" {
			fmt.Fprintf(&report, "/%s %s", field.label, value)
		}
	}

	_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
		Content: report.String(),
	})
	if err != nil {
		log.Printf(
			"Error sending response for /pirep for %s: %v\n",
			helpers.GetMemberName(i.Member),
			err,
		)
	}
}
