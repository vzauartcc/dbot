package helpers

import (
	"testing"

	"github.com/vzauartcc/dbot/internal/api/models"
)

func TestCalculateNewNickname(t *testing.T) {
	cases := []struct {
		user   models.User
		expect string
	}{
		{ // No special roles, rating S1 suffix
			user: models.User{
				FirstName: "John",
				LastName:  "Doe",
				Rating:    2,
			},
			expect: "John Doe | S1",
		},
		{ // ATM and DATM roles, rating SUS not suffix
			user: models.User{
				FirstName: "Jane",
				LastName:  "Doe",
				Rating:    0,
				Roles:     []string{"atm", "datm"},
			},
			expect: "Jane Doe | ATM",
		},
		{ // Home facility zhq gives VATUSA, rating S3 suffix
			user: models.User{
				FirstName:    "Bob",
				LastName:     "Doe",
				Rating:       4,
				HomeFacility: "zhq",
			},
			expect: "Bob Doe | VATUSA",
		},
		{ // I3 role with visitor true gives C1
			user: models.User{
				FirstName: "Eve",
				LastName:  "Doe",
				Rating:    0,
				IsVisitor: true,
				Roles:     []string{"I3"},
			},
			expect: "Eve Doe | C1",
		},
		{ // I3 role with visitor false gives I3
			user: models.User{
				FirstName: "Mallory",
				LastName:  "Doe",
				Rating:    0,
				IsVisitor: false,
				Roles:     []string{"I3"},
			},
			expect: "Mallory Doe | I3",
		},
		{ // ATM and zhq roles together
			user: models.User{
				FirstName: "Peggy",
				LastName:  "Doe",
				Rating:    0,
				Roles:     []string{"atm", "zhq"},
			},
			expect: "Peggy Doe | ATM",
		},
	}

	for _, c := range cases {
		got := calculateNewNickname(c.user)
		if got != c.expect {
			t.Fatalf("for %+v expected %s but got %s", c.user, c.expect, got)
		}
	}
}
