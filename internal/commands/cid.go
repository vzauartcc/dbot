package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var CID = &discordgo.ApplicationCommand{
	Name:                     "cid",
	Description:              "Get a user's CID (if linked)",
	DefaultMemberPermissions: int64Ptr(discordgo.PermissionModerateMembers),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to attempt to retrieve their CID",
			Required:    true,
		},
	},
}

func CidHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "cid") {
		return
	}

	userOption := i.ApplicationCommandData().Options[0]
	targetUser := userOption.UserValue(s)

	user, err := zauapi.GetClient().GetUserByID(targetUser.ID)
	if err != nil {
		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Unable to retrieve user. Their Discord account may not be linked.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /cid for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"%s's CID is **%d**.[Link to profile](https://zauartcc.org/controllers/%d)",
			targetUser.Mention(),
			user.CID,
			user.CID,
		),
		Flags: discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Printf("Error sending success response of /cid for %s: %v\n", i.User.ID, err)
	}
}
