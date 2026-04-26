package handlers

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var reminderMessages map[string]string

func HandleMessage(s *discordgo.Session, message *discordgo.MessageCreate) {
	cfg, ok := models.GetConfig(message.GuildID)
	if !ok {
		return
	}

	if message.Author.ID == s.State.User.ID || message.Author.Bot {
		return
	}

	if _, ok := cfg.GetRepostChannels()[message.ChannelID]; ok {
		handleRepostChannel(s, message, cfg.GetRepostChannels()[message.ChannelID])

		return
	}

	if _, ok := cfg.GetReminderChannels()[message.ChannelID]; ok {
		handleReminderChannel(s, message, cfg.GetReminderChannels()[message.ChannelID])

		return
	}
}

func handleRepostChannel(s *discordgo.Session, message *discordgo.MessageCreate, title string) {
	avatarURL := message.Author.AvatarURL("")
	if message.Member != nil && message.Member.Avatar != "" {
		avatarURL = message.Member.AvatarURL("")
	}

	username := message.Author.Username
	if message.Member != nil {
		username = message.Member.Nick
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: message.Content,
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    username,
			IconURL: avatarURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	attachments := make([]string, 0, len(message.Attachments))
	for _, a := range message.Attachments {
		attachments = append(attachments, a.URL)
	}

	_, err := helpers.ChannelMessageSendComplex(s, message.ChannelID, &discordgo.MessageSend{
		Embeds:  []*discordgo.MessageEmbed{embed},
		Content: strings.Join(attachments, "\n"),
	})
	if err != nil {
		log.Printf("Failed to post embed in repost channel %s: %v\n", message.ChannelID, err)

		return
	}

	err = helpers.ChannelMessageDelete(s, message.ChannelID, message.ID)
	if err != nil {
		log.Printf(
			"Error deleting message %s in report channel %s: %v\n",
			message.ID,
			message.ChannelID,
			err,
		)
	}
}

func handleReminderChannel(s *discordgo.Session, message *discordgo.MessageCreate, content string) {
	if reminderMessages == nil {
		reminderMessages = make(map[string]string)
	}

	if reminderMessages[message.ChannelID] == "" {
		log.Printf("No existing reminder message in %s, creating\n", message.ChannelID)
		sendMessage(s, message.ChannelID, content)

		return
	}

	oldReminderMessage, err := helpers.ChannelMessage(s, message.ChannelID, reminderMessages[message.ChannelID])
	if err != nil {
		log.Printf(
			"Error getting existing reminder message in %s. Sending a new one. %v\n",
			message.ChannelID,
			err,
		)

		sendMessage(s, message.ChannelID, content)

		return
	}

	if time.Since(oldReminderMessage.Timestamp) >= 90*time.Second {
		err = helpers.ChannelMessageDelete(s, oldReminderMessage.ChannelID, oldReminderMessage.ID)
		if err != nil {
			log.Printf("Error deleting old reminder message in %s: %v\n", message.ChannelID, err)
		}

		sendMessage(s, message.ChannelID, content)
	} else {
		log.Printf("Reminder message in %s is less than 90 seconds old: %v\n", message.ChannelID, time.Since(oldReminderMessage.Timestamp))
	}
}

func sendMessage(s *discordgo.Session, channelID, content string) {
	msg, err := helpers.ChannelMessageSend(s, channelID, content)
	if err != nil {
		log.Printf("Error sending reminder message in %s: %v\n", channelID, err)
		return
	}

	reminderMessages[msg.ChannelID] = msg.ID
}
