package handlers

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

func handleMessageComponents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "grantrole") {
		return
	}

	cfg, ok := models.GetConfig(i.GuildID)
	if !ok {
		return
	}

	data := i.MessageComponentData()

	if !strings.HasPrefix(data.CustomID, "grantrole:") {
		return
	}

	parts := strings.Split(data.CustomID, ":")
	if len(parts) != 2 {
		log.Printf("Invalid grantrole custom ID: %s\n", data.CustomID)
		return
	}

	targetUserID := parts[1]
	targetRole := data.Values[0]

	canGrant := false

	for _, role := range i.Member.Roles {
		grantable, ok := cfg.GetRoleGrants()[role]
		if !ok {
			continue
		}

		if slices.Contains(grantable, targetRole) {
			canGrant = true
		}
	}

	if !canGrant {
		_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: "You do not have permission to grant this role.",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf("Error sending role grant failure for %s: %v\n", i.User.ID, err)
		}

		return
	}

	err := helpers.GuildMemberRoleAdd(s, i.GuildID, targetUserID, targetRole)
	if err != nil {
		log.Printf("Error granting role %s to %s: %v\n", targetRole, targetUserID, err)

		_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Error granting role %s to %s: %v", targetRole, targetUserID, err),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf("Error sending role grant failure for %s: %v\n", i.User.ID, err)
		}

		return
	}

	_, err = helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Granted role %s to %s", targetRole, targetUserID),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Printf("Error sending role grant success for %s: %v\n", i.User.ID, err)
	}
}
