package helpers

import (
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
)

func SetNickname(s *discordgo.Session, member *discordgo.Member, user models.User) error {
	var err error

	newNick := calculateNewNickname(user)

	if newNick != member.Nick {
		log.Printf("Updating nickname for %s to: \"%s\"\n", GetMemberName(member), newNick)

		err = s.GuildMemberNickname(member.GuildID, member.User.ID, newNick)
	}

	return err
}

func calculateNewNickname(user models.User) string {
	newNick := user.FirstName + " " + user.LastName

	roles := make([]string, 0)
	roles = append(roles, user.Roles...)

	if user.Rating >= 0 && user.Rating <= len(ratingsToString) {
		roles = append(roles, ratingsToString[user.Rating])
	}

	if strings.ToLower(user.HomeFacility) == "zhq" {
		roles = append(roles, "zhq")
	}

	switch {
	case slices.Contains(roles, "atm"):
		newNick += " | ATM"
	case slices.Contains(roles, "datm"):
		newNick += " | DATM"
	// case slices.Contains(roles, "ta"):
	// 	newNick += " | TA"
	// case slices.Contains(roles, "ec"):
	// 	newNick += " | EC"
	// case slices.Contains(roles, "fe"):
	// 	newNick += " | FE"
	// case slices.Contains(roles, "WM"):
	// 	newNick += " | WM"
	case slices.Contains(roles, "zhq"):
		newNick += " | VATUSA"
	case slices.Contains(roles, "ins"):
		newNick += " | INS"
	case slices.Contains(roles, "mtr"):
		newNick += " | MTR"
	case slices.Contains(roles, "ia"):
		newNick += " | IA"
	case slices.Contains(roles, "ADM"):
		newNick += " | ADM"
	case slices.Contains(roles, "SUP"):
		newNick += " | SUP"
	case slices.Contains(roles, "I3"):
		if user.IsVisitor {
			newNick += " | C1"
		} else {
			newNick += " | I3"
		}
	case slices.Contains(roles, "I1"):
		if user.IsVisitor {
			newNick += " | C1"
		} else {
			newNick += " | I1"
		}
	case slices.Contains(roles, "C3"):
		newNick += " | C3"
	case slices.Contains(roles, "C1"):
		newNick += " | C1"
	case slices.Contains(roles, "S3"):
		newNick += " | S3"
	case slices.Contains(roles, "S2"):
		newNick += " | S2"
	case slices.Contains(roles, "S1"):
		newNick += " | S1"
	case slices.Contains(roles, "OBS"):
		newNick += " | OBS"
	}

	return newNick
}
