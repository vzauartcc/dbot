package commands

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var CreateInstructorChannel = &discordgo.ApplicationCommand{
	Name:        "create-instructor-channel",
	Description: "Create a training coordination channel for an instructor",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The instructor the channel is for",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "cid",
			Description: "The instructor's CID",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "operating-initials",
			Description: "The instructor's Operating Initials",
			Required:    true,
			MinLength:   new(2),
			MaxLength:   2,
		},
	},
}

func CreateInstructorChannelHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "create-instructor-channel", true) {
		return
	}

	// Training Administrator or Senior Staff role only.
	if !slices.Contains(i.Member.Roles, "1215950778120933467") &&
		!slices.Contains(i.Member.Roles, "485491894168584202") {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "You do not have permission to run this command.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	userOption := i.ApplicationCommandData().GetOption("user")
	if userOption == nil {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "User is required",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	targetUser := userOption.UserValue(s)

	targetInstructor, err := helpers.GuildMember(s, i.GuildID, targetUser.ID)
	if err != nil {
		_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Unable to retrieve member.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	inputOI := i.ApplicationCommandData().GetOption("operating-initials")
	if inputOI == nil {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Operating initials are required",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	instructorOI := inputOI.StringValue()

	if instructorOI == "" || len(instructorOI) != 2 {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Invalid operating initials",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	inputCID := i.ApplicationCommandData().GetOption("cid")
	if inputCID == nil {
		_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "CID is required",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	instructorCID := inputCID.IntValue()

	// Founder CIDs start at 800001, not possible to be less than a founder.
	if instructorCID < 800000 {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Invalid CID",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	instructorName := strings.Fields(strings.Split(targetInstructor.Nick, "|")[0])

	if len(instructorName) < 2 {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Instructor nickname is invalid, fix and try again",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	createdRole, err := helpers.GuildRoleCreate(s, i.GuildID, &discordgo.RoleParams{
		Name:        fmt.Sprintf("%s Student (%d)", instructorOI, instructorCID),
		Hoist:       new(false),
		Mentionable: new(true),
	})
	if err != nil {
		log.Printf(
			"Error creating instructor student role for %s: %v\n",
			targetInstructor.Nick,
			err,
		)

		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error creating instructor role",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return
	}

	createdChannel := createChannel(
		s,
		i,
		targetInstructor,
		createdRole,
		instructorCID,
		instructorName,
	)
	if createdChannel == nil {
		return
	}

	log.Println("Created instructor channel", createdChannel.ID)

	_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Successfully created <@&%s> and <#%s>!",
			createdRole.ID,
			createdChannel.ID,
		),
		Flags: discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Printf(
			"Error sending success response of /create-instructor-channel for %s: %v\n",
			i.User.ID,
			err,
		)
	}
}

func createChannel(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	instructor *discordgo.Member,
	role *discordgo.Role,
	instructorCID int64,
	instructorName []string,
) *discordgo.Channel {
	createdChannel, err := helpers.GuildChannelCreateComplex(s,
		i.GuildID,
		discordgo.GuildChannelCreateData{
			Name: fmt.Sprintf(
				"%s-%s-training",
				instructorName[0],
				string([]rune(instructorName[1])[0]),
			),
			Type: discordgo.ChannelTypeGuildText,
			Topic: fmt.Sprintf(
				"Training coordination channel for %s. cid:%d id:%s",
				instructorName[0],
				instructorCID,
				instructor.User.ID,
			),
			// Gets past most of the announcment and other channels.
			Position: 13,
			PermissionOverwrites: []*discordgo.PermissionOverwrite{
				// Deny @everyone, which uses the guild ID for the role ID.
				{
					ID:    i.GuildID,
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: 0,
					Deny:  discordgo.PermissionViewChannel,
				},
				// Instructor user.
				{
					ID:    instructor.User.ID,
					Type:  discordgo.PermissionOverwriteTypeMember,
					Allow: discordgo.PermissionViewChannel | discordgo.PermissionMentionEveryone,
					Deny:  0,
				},
				// Student role.
				{
					ID:    role.ID,
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
				// Training Administrator role.
				{
					ID:    "1215950778120933467",
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel | discordgo.PermissionMentionEveryone | discordgo.PermissionManageMessages | discordgo.PermissionManageThreads,
					Deny:  0,
				},
				// Senior Staff role.
				{
					ID:    "485491894168584202",
					Type:  discordgo.PermissionOverwriteTypeRole,
					Allow: discordgo.PermissionViewChannel,
					Deny:  0,
				},
			},
			ParentID: "485494987853922306",
			NSFW:     false,
		},
	)
	if err != nil {
		log.Printf("Error creating instructor text channel for %s: %v\n", instructor.Nick, err)

		err := helpers.GuildRoleDelete(s, i.GuildID, role.ID)
		if err != nil {
			log.Printf("Error deleting %s role: %v\n", role.Name, err)
		}

		_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to create channel",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf(
				"Error sending failure response of /create-instructor-channel for %s: %v\n",
				helpers.GetMemberName(i.Member),
				err,
			)
		}

		return nil
	}

	return createdChannel
}
