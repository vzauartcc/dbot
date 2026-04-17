package commands

import (
	"errors"
	"log"

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

func GetRolesHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "getroles") {
		return
	}

	cfg, ok := models.GetConfig(i.GuildID)
	if !ok {
		log.Printf(
			"%s used the /getroles command in an unsupported guild: %s",
			helpers.GetMemberName(i.Member),
			i.GuildID,
		)

		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "This server is not configured for automated role handling.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending /getroles failure for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	user, err := zauapi.GetClient().GetUserByID(i.User.ID)
	if err != nil {
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error verifying account. Ensure your account is linked at [zauartcc.org](https://zauartcc.org/dash)",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending /getroles failure for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	rolesToGive := helpers.RolesToAdd(cfg, user)

	errs := helpers.ExchangeRoles(s, i.Member, cfg, rolesToGive, "/getroles Command")
	if len(errs) != 0 {
		log.Printf("Error processing /getroles for %s: %v\n", i.User.ID, errs)

		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error updating your roles.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure of /getroles for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}
	} else {
		_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
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
