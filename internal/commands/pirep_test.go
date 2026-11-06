package commands

import (
	"errors"
	"maps"
	"slices"
	"testing"
)

func TestParsePirepLocation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []pirepFix
	}{
		{
			name:  "fix only",
			input: "ABC",
			want:  []pirepFix{{Ident: "ABC"}},
		},
		{
			name:  "fix with bearing and distance",
			input: "ABC 090025",
			want:  []pirepFix{{Ident: "ABC", Bearing: 90, Distance: 25, HasBearingDistance: true}},
		},
		{
			name:  "lowercase input is normalized",
			input: "abc 090025",
			want:  []pirepFix{{Ident: "ABC", Bearing: 90, Distance: 25, HasBearingDistance: true}},
		},
		{
			name:  "fix to fix",
			input: "ABC-DEF",
			want:  []pirepFix{{Ident: "ABC"}, {Ident: "DEF"}},
		},
		{
			name:  "fix to fix with trailing bearing and distance",
			input: "ABC-DEF 120020",
			want: []pirepFix{
				{Ident: "ABC"},
				{Ident: "DEF", Bearing: 120, Distance: 20, HasBearingDistance: true},
			},
		},
		{
			name:  "bearing and distance from first fix",
			input: "ABC 045020-DEF 120005",
			want: []pirepFix{
				{Ident: "ABC", Bearing: 45, Distance: 20, HasBearingDistance: true},
				{Ident: "DEF", Bearing: 120, Distance: 5, HasBearingDistance: true},
			},
		},
		{
			name:  "three fixes",
			input: "ABC-DEF-GHI",
			want:  []pirepFix{{Ident: "ABC"}, {Ident: "DEF"}, {Ident: "GHI"}},
		},
		{
			name:  "bearing and distance attached to ident",
			input: "ABC090025-DEF120005",
			want: []pirepFix{
				{Ident: "ABC", Bearing: 90, Distance: 25, HasBearingDistance: true},
				{Ident: "DEF", Bearing: 120, Distance: 5, HasBearingDistance: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePirepLocation(tt.input)
			if err != nil {
				t.Fatalf("parsePirepLocation(%q) returned error: %v", tt.input, err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("parsePirepLocation(%q) = %+v; want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePirepLocationErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "short ident", input: "AB"},
		{name: "ident with symbol", input: "A$C"},
		{name: "bearing out of range", input: "ABC 361000"},
		{name: "short bearing and distance", input: "ABC 09002"},
		{name: "bearing and distance without a fix", input: "090025"},
		{name: "nondigit bearing and distance", input: "ABC09002X"},
		{name: "repeated bearing and distance", input: "ABC 090025 120005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePirepLocation(tt.input)
			if !errors.Is(err, errPirepLocation) {
				t.Errorf("parsePirepLocation(%q) error = %v; want %v", tt.input, err, errPirepLocation)
			}
		})
	}
}

func TestFormatPirepLocation(t *testing.T) {
	tests := []struct {
		name  string
		fixes []pirepFix
		want  string
	}{
		{
			name:  "single fix",
			fixes: []pirepFix{{Ident: "ABC"}},
			want:  "ABC",
		},
		{
			name:  "fix with padding",
			fixes: []pirepFix{{Ident: "ABC", Bearing: 45, Distance: 20, HasBearingDistance: true}},
			want:  "ABC 045020",
		},
		{
			name: "chain",
			fixes: []pirepFix{
				{Ident: "ABC", Bearing: 45, Distance: 20, HasBearingDistance: true},
				{Ident: "DEF", Bearing: 120, Distance: 5, HasBearingDistance: true},
				{Ident: "GHI"},
			},
			want: "ABC 045020-DEF 120005-GHI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPirepLocation(tt.fixes)
			if got != tt.want {
				t.Errorf("formatPirepLocation(%+v) = %q; want %q", tt.fixes, got, tt.want)
			}
		})
	}
}

func TestParsePirepTime(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantHour   int
		wantMinute int
	}{
		{name: "midnight", input: "0000", wantHour: 0, wantMinute: 0},
		{name: "morning", input: "0915", wantHour: 9, wantMinute: 15},
		{name: "last minute of day", input: "2359", wantHour: 23, wantMinute: 59},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hour, minute, err := parsePirepTime(tt.input)
			if err != nil {
				t.Fatalf("parsePirepTime(%q) returned error: %v", tt.input, err)
			}

			if hour != tt.wantHour || minute != tt.wantMinute {
				t.Errorf(
					"parsePirepTime(%q) = %02d%02d; want %02d%02d",
					tt.input,
					hour,
					minute,
					tt.wantHour,
					tt.wantMinute,
				)
			}
		})
	}
}

func TestParsePirepTimeErrors(t *testing.T) {
	for _, input := range []string{"", "915", "09150", "24:00", "2400", "0960", "09x5"} {
		t.Run(input, func(t *testing.T) {
			_, _, err := parsePirepTime(input)
			if !errors.Is(err, errPirepTime) {
				t.Errorf("parsePirepTime(%q) error = %v; want %v", input, err, errPirepTime)
			}
		})
	}
}

func TestParsePirepFlightLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantAlt     int
		wantUnknown bool
	}{
		{name: "low level", input: "095", wantAlt: 9500},
		{name: "flight level", input: "310", wantAlt: 31000},
		{name: "unknown lowercase", input: "unkn", wantUnknown: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			altitude, unknown, err := parsePirepFlightLevel(tt.input)
			if err != nil {
				t.Fatalf("parsePirepFlightLevel(%q) returned error: %v", tt.input, err)
			}

			if altitude != tt.wantAlt || unknown != tt.wantUnknown {
				t.Errorf(
					"parsePirepFlightLevel(%q) = %d (unknown %t); want %d (unknown %t)",
					tt.input,
					altitude,
					unknown,
					tt.wantAlt,
					tt.wantUnknown,
				)
			}
		})
	}
}

func TestParsePirepFlightLevelErrors(t *testing.T) {
	for _, input := range []string{"", "95", "0950", "0A5", "UNK"} {
		t.Run(input, func(t *testing.T) {
			_, _, err := parsePirepFlightLevel(input)
			if !errors.Is(err, errPirepFlightLevel) {
				t.Errorf(
					"parsePirepFlightLevel(%q) error = %v; want %v",
					input,
					err,
					errPirepFlightLevel,
				)
			}
		})
	}
}

func TestParsePirepAircraftType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "lowercase is normalized", input: "b737", want: "B737"},
		{name: "three characters", input: "PA32", want: "PA32"},
		{name: "unknown", input: "unkn", want: "UNKN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePirepAircraftType(tt.input)
			if err != nil {
				t.Fatalf("parsePirepAircraftType(%q) returned error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("parsePirepAircraftType(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePirepAircraftTypeErrors(t *testing.T) {
	for _, input := range []string{"", "B7377", "B-73", "B 73"} {
		t.Run(input, func(t *testing.T) {
			_, err := parsePirepAircraftType(input)
			if !errors.Is(err, errPirepAircraftType) {
				t.Errorf(
					"parsePirepAircraftType(%q) error = %v; want %v",
					input,
					err,
					errPirepAircraftType,
				)
			}
		})
	}
}

func TestParsePirepSky(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []pirepSkyLayer
	}{
		{
			name:  "base and cover",
			input: "038 BKN",
			want:  []pirepSkyLayer{{Base: "038", Cover: "BKN"}},
		},
		{
			name:  "base cover and tops",
			input: "038 OVC 045",
			want:  []pirepSkyLayer{{Base: "038", Cover: "OVC", Tops: "045"}},
		},
		{
			name:  "two layers",
			input: "055 SCT 073/085 BKN 105",
			want: []pirepSkyLayer{
				{Base: "055", Cover: "SCT", Tops: "073"},
				{Base: "085", Cover: "BKN", Tops: "105"},
			},
		},
		{
			name:  "unknown base",
			input: "UNKN OVC",
			want:  []pirepSkyLayer{{Base: "UNKN", Cover: "OVC"}},
		},
		{
			name:  "lowercase is normalized",
			input: "025 ovc 095",
			want:  []pirepSkyLayer{{Base: "025", Cover: "OVC", Tops: "095"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePirepSky(tt.input)
			if err != nil {
				t.Fatalf("parsePirepSky(%q) returned error: %v", tt.input, err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("parsePirepSky(%q) = %+v; want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePirepSkyErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "cover only", input: "BKN"},
		{name: "unknown cover", input: "038 XYZ"},
		{name: "short base", input: "38 BKN"},
		{name: "nondigit tops", input: "038 BKN 04X"},
		{name: "too many fields", input: "038 BKN 045 060"},
		{name: "empty layer", input: "038 BKN/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePirepSky(tt.input)
			if !errors.Is(err, errPirepSky) {
				t.Errorf("parsePirepSky(%q) error = %v; want %v", tt.input, err, errPirepSky)
			}
		})
	}
}

func TestFormatPirepSky(t *testing.T) {
	layers := []pirepSkyLayer{
		{Base: "055", Cover: "SCT", Tops: "073"},
		{Base: "UNKN", Cover: "OVC"},
	}

	got := formatPirepSky(layers)
	want := "055 SCT 073/UNKN OVC"

	if got != want {
		t.Errorf("formatPirepSky(%+v) = %q; want %q", layers, got, want)
	}
}

func TestParsePirepTemperature(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "positive", input: "15", want: 15},
		{name: "negative", input: "-06", want: -6},
		{name: "single digit", input: "4", want: 4},
		{name: "freezing", input: "0", want: 0},
		{name: "whitespace", input: " -21 ", want: -21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePirepTemperature(tt.input)
			if err != nil {
				t.Fatalf("parsePirepTemperature(%q) returned error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("parsePirepTemperature(%q) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePirepTemperatureErrors(t *testing.T) {
	for _, input := range []string{"", "-", "+5", "15C", "1000", "-123", "1.5"} {
		t.Run(input, func(t *testing.T) {
			_, err := parsePirepTemperature(input)
			if !errors.Is(err, errPirepTemperature) {
				t.Errorf(
					"parsePirepTemperature(%q) error = %v; want %v",
					input,
					err,
					errPirepTemperature,
				)
			}
		})
	}
}

func TestParsePirepWind(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantDirection int
		wantSpeed     int
	}{
		{name: "wind", input: "270045", wantDirection: 270, wantSpeed: 45},
		{name: "strong wind", input: "280110", wantDirection: 280, wantSpeed: 110},
		{name: "calm", input: "000000", wantDirection: 0, wantSpeed: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction, speed, err := parsePirepWind(tt.input)
			if err != nil {
				t.Fatalf("parsePirepWind(%q) returned error: %v", tt.input, err)
			}

			if direction != tt.wantDirection || speed != tt.wantSpeed {
				t.Errorf(
					"parsePirepWind(%q) = %03d/%03d; want %03d/%03d",
					tt.input,
					direction,
					speed,
					tt.wantDirection,
					tt.wantSpeed,
				)
			}
		})
	}
}

func TestParsePirepWindErrors(t *testing.T) {
	for _, input := range []string{"", "27045", "2700450", "361000", "27X045"} {
		t.Run(input, func(t *testing.T) {
			_, _, err := parsePirepWind(input)
			if !errors.Is(err, errPirepWind) {
				t.Errorf("parsePirepWind(%q) error = %v; want %v", input, err, errPirepWind)
			}
		})
	}
}

func TestBuildPirep(t *testing.T) {
	tests := []struct {
		name string
		opts map[string]string
		want string
	}{
		{
			name: "routine report",
			opts: map[string]string{
				"type": "UA",
				"ov":   "RFD 170030",
				"tm":   "1315",
				"fl":   "160",
				"tp":   "PA60",
				"sk":   "025 OVC 095/180 OVC",
				"ta":   "-21",
				"wv":   "270048",
			},
			want: "**PIREP**\n`UA /OV RFD 170030/TM 1315/FL160/TP PA60" +
				"/SK 025 OVC 095/180 OVC/TA -21/WV 270048`",
		},
		{
			name: "urgent report with remarks",
			opts: map[string]string{
				"type": "UUA",
				"ov":   "DHT 360015-AMA-CDS",
				"tm":   "2116",
				"fl":   "050",
				"tp":   "PA32",
				"sk":   "UNKN OVC",
				"wx":   "FV03 R",
				"ta":   "04",
				"tb":   "LGT",
				"rm":   "HVY RAIN",
			},
			want: "**PIREP**\n`UUA /OV DHT 360015-AMA-CDS/TM 2116/FL050/TP PA32" +
				"/SK UNKN OVC/WX FV03 R/TA 04/TB LGT/RM HVY RAIN`",
		},
		{
			name: "unknown flight level and aircraft type",
			opts: map[string]string{
				"type": "ua",
				"ov":   "abc",
				"tm":   "0000",
				"fl":   "unkn",
				"tp":   "unkn",
				"ic":   "LGT RIME",
				"ta":   "-06",
			},
			want: "**PIREP**\n`UA /OV ABC/TM 0000/FLUNKN/TP UNKN/TA -06/IC LGT RIME`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildPirep(tt.opts)
			if err != nil {
				t.Fatalf("buildPirep() returned error: %v", err)
			}

			if got != tt.want {
				t.Errorf("buildPirep() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestBuildPirepErrors(t *testing.T) {
	tests := []struct {
		name     string
		override map[string]string
		want     error
	}{
		{name: "invalid type", override: map[string]string{"type": "XX"}, want: errPirepType},
		{name: "invalid location", override: map[string]string{"ov": "AB"}, want: errPirepLocation},
		{name: "invalid time", override: map[string]string{"tm": "2500"}, want: errPirepTime},
		{name: "invalid flight level", override: map[string]string{"fl": "95"}, want: errPirepFlightLevel},
		{name: "invalid aircraft type", override: map[string]string{"tp": "B7377"}, want: errPirepAircraftType},
		{name: "invalid sky", override: map[string]string{"sk": "038 XYZ"}, want: errPirepSky},
		{name: "invalid temperature", override: map[string]string{"ta": "hot"}, want: errPirepTemperature},
		{name: "invalid wind", override: map[string]string{"wv": "27045"}, want: errPirepWind},
		{
			name:     "icing without temperature",
			override: map[string]string{"ic": "LGT RIME"},
			want:     errPirepIcingTemperature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := validPirepOptions()
			maps.Copy(opts, tt.override)

			_, err := buildPirep(opts)
			if !errors.Is(err, tt.want) {
				t.Errorf("buildPirep() error = %v; want %v", err, tt.want)
			}
		})
	}
}

func validPirepOptions() map[string]string {
	return map[string]string{
		"type": "UA",
		"ov":   "ABC",
		"tm":   "0915",
		"fl":   "095",
		"tp":   "B737",
	}
}
