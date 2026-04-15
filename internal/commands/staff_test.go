package commands

import (
	"testing"

	zauapi "github.com/vzauartcc/dbot/internal/api"
)

func TestGenerateStaffEmbed(t *testing.T) {
	staff := zauapi.Staff{
		ATM: zauapi.StaffPosition{
			Users: []zauapi.User{
				{FirstName: "John", LastName: "Doe"},
				{FirstName: "Jane", LastName: "Smith"},
			},
		},
		DATM: zauapi.StaffPosition{Users: []zauapi.User{{FirstName: "Bob", LastName: "Brown"}}},
		TA:   zauapi.StaffPosition{Users: []zauapi.User{{FirstName: "Alice", LastName: "Green"}}},
		EC:   zauapi.StaffPosition{Users: []zauapi.User{{FirstName: "Charlie", LastName: "White"}}},
		FE:   zauapi.StaffPosition{Users: []zauapi.User{{FirstName: "David", LastName: "Black"}}},
		WM:   zauapi.StaffPosition{Users: []zauapi.User{{FirstName: "Eve", LastName: "Gray"}}},
	}

	embed := generateStaffEmbed(staff)

	if embed.Title != "ZAU Staff" {
		t.Errorf("unexpected title %s", embed.Title)
	}

	if len(embed.Fields) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(embed.Fields))
	}

	for _, field := range embed.Fields {
		switch field.Name {
		case "Air Traffic Manager":
			expected := "John Doe, Jane Smith [Email](mailto:atm@zauartcc.org)"
			if field.Value != expected {
				t.Errorf("unexpected value %s", field.Value)
			}
		case "Deputy Air Traffic Manager":
			expected := "Bob Brown [Email](mailto:datm@zauartcc.org)"
			if field.Value != expected {
				t.Errorf("unexpected value %s", field.Value)
			}
		case "Training Administrator":
			expected := "Alice Green [Email](mailto:ta@zauartcc.org)"
			if field.Value != expected {
				t.Errorf("unexpected value %s", field.Value)
			}
		case "Event Coordinator":
			expected := "Charlie White [Email](mailto:events@zauartcc.org)"
			if field.Value != expected {
				t.Errorf("unexpected value %s", field.Value)
			}
		case "Facility Engineer":
			expected := "David Black [Email](mailto:facilities@zauartcc.org)"
			if field.Value != expected {
				t.Errorf("expected value %s", field.Value)
			}
		case "Web Team":
			expected := "Eve Gray [Email](mailto:wm@zauartcc.org)"
			if field.Value != expected {
				t.Errorf("unexpected value %s", field.Value)
			}
		default:
			t.Fatalf("unknown field name %s", field.Name)
		}
	}
}
