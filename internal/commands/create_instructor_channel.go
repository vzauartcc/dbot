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
		sendFollowUpMessage(
			s,
			i.Interaction,
			"create-instructor-channel",
			"You do not have permission to run this command.",
		)

		return
	}

	userOption := i.ApplicationCommandData().GetOption("user")
	if userOption == nil {
		sendFollowUpMessage(s, i.Interaction, "create-instructor-channel", "User is required.")

		return
	}

	targetUser := userOption.UserValue(s)

	targetInstructor, err := helpers.GuildMember(s, i.GuildID, targetUser.ID)
	if err != nil {
		sendFollowUpMessage(
			s,
			i.Interaction,
			"create-instructor-channel",
			"Unable to retrieve member.",
		)

		return
	}

	inputOI := i.ApplicationCommandData().GetOption("operating-initials")
	if inputOI == nil {
		sendFollowUpMessage(
			s,
			i.Interaction,
			"create-instructor-channel",
			"Operating initials are required.",
		)

		return
	}

	instructorOI := inputOI.StringValue()

	if instructorOI == "" || len(instructorOI) != 2 {
		sendFollowUpMessage(
			s,
			i.Interaction,
			"create-instructor-channel",
			"Invalid instructor operating initials.",
		)

		return
	}

	inputCID := i.ApplicationCommandData().GetOption("cid")
	if inputCID == nil {
		sendFollowUpMessage(s, i.Interaction, "create-instructor-channel", "CID is required.")

		return
	}

	instructorCID := inputCID.IntValue()

	// Founder CIDs start at 800001, not possible to be less than a founder.
	if instructorCID < 800000 {
		sendFollowUpMessage(s, i.Interaction, "create-instructor-channel", "Invalid CID.")

		return
	}

	instructorName := strings.Fields(strings.Split(targetInstructor.Nick, "|")[0])

	if len(instructorName) < 2 {
		sendFollowUpMessage(
			s,
			i.Interaction,
			"create-instructor-channel",
			"Instructor nickname is invalid, please try again.",
		)

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

		sendFollowUpMessage(
			s,
			i.Interaction,
			"create-instructor-channel",
			"Error creating instructor role",
		)

		return
	}

	createdChannel := createChannel(
		s,
		i.Interaction,
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
	i *discordgo.Interaction,
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
			// Gets past most of the announcement and other channels.
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

		sendFollowUpMessage(
			s,
			i,
			"create-instructor-channel",
			"Failed to create instructor coordination channel.",
		)

		return nil
	}

	// Ensure channel is sorted correctly, since discord sometimes ignores
	// the Position field in the GuildChannelCreateComplex call.
	_, err = helpers.ChannelEdit(s, createdChannel.ID, &discordgo.ChannelEdit{
		Position: new(13),
	})
	if err != nil {
		log.Printf(
			"Error sorting newly created instructor text channel %s: %v\n",
			instructor.Nick,
			err,
		)
	}

	return createdChannel
}
