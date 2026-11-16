package commands

import (
	"errors"
	"log"
	"slices"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var ErrUserNotFound = errors.New("user not found")

var GetRoles = &discordgo.ApplicationCommand{
	Name:        "getroles",
	Description: "Get roles",
}

var GiveRoles = &discordgo.ApplicationCommand{
	Name:        "giveroles",
	Description: "Get roles",
}

func GetRolesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "getroles", false) {
		return
	}

	cfg, ok := models.GetConfig(i.GuildID)
	if !ok {
		log.Printf(
			"%s used the /getroles command in an unsupported guild: %s",
			helpers.GetMemberName(i.Member),
			i.GuildID,
		)

		sendFollowUpMessage(
			s,
			i.Interaction,
			"getroles",
			"This server is not configured for automated role management.",
		)

		return
	}

	user, err := zauapi.GetClient().GetUserByID(i.Member.User.ID)
	if err != nil {
		sendFollowUpMessage(
			s,
			i.Interaction,
			"getroles",
			"Error verifying your account. Please ensure your account is linked at [zauartcc.org](https://zauartcc.org/dash).",
		)

		return
	}

	rolesToGive := helpers.RolesToAdd(cfg, user)

	errs := helpers.ExchangeRoles(s, i.Member, cfg, rolesToGive, "/getroles Command")
	if len(errs) != 0 {
		log.Printf("Error processing /getroles for %s: %v\n", i.Member.User.ID, errs)

		if slices.Contains(errs, helpers.ErrNoChanges) {
			sendFollowUpMessage(s, i.Interaction, "getroles", "Roles synced! (No changes)")
		} else {
			sendFollowUpMessage(
				s,
				i.Interaction,
				"getroles",
				"An error occurred updating your roles.",
			)
		}
	} else {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Roles updated!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending success of /getroles for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		err = helpers.SetNickname(s, i.Member, user)
		if err != nil {
			log.Printf("Error updating nickname of %s: %v\n", helpers.GetMemberName(i.Member), err)
		}
	}
}
