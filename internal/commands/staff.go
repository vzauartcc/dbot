package commands

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var Staff = &discordgo.ApplicationCommand{
	Name:        "staff",
	Description: "List staff",
}

func StaffHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "staff") {
		return
	}

	staff, err := zauapi.GetClient().GetStaff()
	if err != nil {
		log.Printf("Error getting staff members: %v\n", err)

		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "An error has occurred.",
		})
		if err != nil {
			log.Printf(
				"Error sending success response for /staff for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	embed := generateStaffEmbed(staff)

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		log.Printf(
			"Error sending success response for /staff for %s: %v\n",
			helpers.GetMemberName(i.Member),
			err,
		)
	}
}

func generateStaffEmbed(staff models.Staff) *discordgo.MessageEmbed {
	atms := make([]string, 0, len(staff.ATM.Users))
	datms := make([]string, 0, len(staff.DATM.Users))
	tas := make([]string, 0, len(staff.TA.Users))
	ecs := make([]string, 0, len(staff.EC.Users))
	fes := make([]string, 0, len(staff.FE.Users))
	wms := make([]string, 0, len(staff.WM.Users))

	for _, user := range staff.ATM.Users {
		atms = append(atms, user.FirstName+" "+user.LastName)
	}

	for _, user := range staff.DATM.Users {
		datms = append(datms, user.FirstName+" "+user.LastName)
	}

	for _, user := range staff.TA.Users {
		tas = append(tas, user.FirstName+" "+user.LastName)
	}

	for _, user := range staff.EC.Users {
		ecs = append(ecs, user.FirstName+" "+user.LastName)
	}

	for _, user := range staff.FE.Users {
		fes = append(fes, user.FirstName+" "+user.LastName)
	}

	for _, user := range staff.WM.Users {
		wms = append(wms, user.FirstName+" "+user.LastName)
	}

	return &discordgo.MessageEmbed{
		Title:       "ZAU Staff",
		Description: "Here are our staff members",
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Air Traffic Manager",
				Value:  strings.Join(atms, ", ") + " [Email](mailto:atm@zauartcc.org)",
				Inline: false,
			},
			{
				Name:   "Deputy Air Traffic Manager",
				Value:  strings.Join(datms, ", ") + " [Email](mailto:datm@zauartcc.org)",
				Inline: false,
			},
			{
				Name:   "Training Administrator",
				Value:  strings.Join(tas, ", ") + " [Email](mailto:ta@zauartcc.org)",
				Inline: false,
			},
			{
				Name:   "Event Coordinator",
				Value:  strings.Join(ecs, ", ") + " [Email](mailto:events@zauartcc.org)",
				Inline: false,
			},
			{
				Name:   "Facility Engineer",
				Value:  strings.Join(fes, ", ") + " [Email](mailto:facilities@zauartcc.org)",
				Inline: false,
			},
			{
				Name:   "Web Team",
				Value:  strings.Join(wms, ", ") + " [Email](mailto:wm@zauartcc.org)",
				Inline: false,
			},
		},
	}
}
