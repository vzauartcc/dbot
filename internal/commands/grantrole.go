package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var GrantRole = &discordgo.ApplicationCommand{
	Name:        "grantrole",
	Description: "Grant a role to a user",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to grant a role",
			Required:    true,
		},
	},
}

func GrantRoleHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "grantrole") {
		return
	}

	userOption := i.ApplicationCommandData().Options[0]
	targetUser := userOption.UserValue(s)

	cfg, ok := models.GetConfig(i.GuildID)
	if !ok {
		log.Printf(
			"%s used the /grantrole command in an unsupported guild: %s",
			helpers.GetMemberName(i.Member),
			i.GuildID,
		)

		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "This server is not configured for granting roles.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending /grantrole failure for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	roleOptions := make([]discordgo.SelectMenuOption, 0)

	for _, role := range i.Member.Roles {
		if b, ok := cfg.GetRoleGrants()[role]; ok {
			for _, r := range b {
				rl, err := s.State.Role(i.GuildID, r)
				if err == nil {
					roleOptions = append(roleOptions, discordgo.SelectMenuOption{
						Label: rl.Name,
						Value: r,
					})
				}
			}
		}
	}

	if len(roleOptions) == 0 {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "You do not have permission to grant roles to this user.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending /grantrole failure for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	customID := "grantrole-" + targetUser.ID

	_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
		Content: "Select a role to grant to **" + targetUser.Username + "**",
		Flags:   discordgo.MessageFlagsEphemeral,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    customID,
						Placeholder: "Choose a role...",
						MinValues:   nil,
						MaxValues:   1,
						Options:     roleOptions,
					},
				},
			},
		},
	})
	if err != nil {
		log.Printf("Error sending /grantrole response for %s: %v\n", i.User.ID, err)
	}
}
