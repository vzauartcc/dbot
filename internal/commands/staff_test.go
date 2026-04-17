package commands

import (
	"testing"

	"github.com/vzauartcc/dbot/internal/api/models"
)

func TestGenerateStaffEmbed(t *testing.T) {
	staff := models.Staff{
		ATM: models.StaffPosition{
			Users: []models.User{
				{FirstName: "John", LastName: "Doe"},
				{FirstName: "Jane", LastName: "Smith"},
			},
		},
		DATM: models.StaffPosition{Users: []models.User{{FirstName: "Bob", LastName: "Brown"}}},
		TA:   models.StaffPosition{Users: []models.User{{FirstName: "Alice", LastName: "Green"}}},
		EC:   models.StaffPosition{Users: []models.User{{FirstName: "Charlie", LastName: "White"}}},
		FE:   models.StaffPosition{Users: []models.User{{FirstName: "David", LastName: "Black"}}},
		WM:   models.StaffPosition{Users: []models.User{{FirstName: "Eve", LastName: "Gray"}}},
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
