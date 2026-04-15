package handlers

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
)

var reminderMessages map[string]string

func HandleMessage(s *discordgo.Session, message *discordgo.MessageCreate) {
	cfg, ok := zauapi.GetConfig(message.GuildID)
	if !ok {
		return
	}

	if message.Author.ID == s.State.User.ID {
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
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: message.Content,
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    message.Member.Nick,
			IconURL: message.Member.AvatarURL(""),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	attachments := make([]string, 0, len(message.Attachments))
	for _, a := range message.Attachments {
		attachments = append(attachments, a.URL)
	}

	_, err := s.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
		Embeds:  []*discordgo.MessageEmbed{embed},
		Content: strings.Join(attachments, "\n"),
	})
	if err != nil {
		log.Printf("Failed to post embed in repost channel %s: %v\n", message.ChannelID, err)

		return
	}

	err = s.ChannelMessageDelete(message.ChannelID, message.ID)
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
		sendMessage(s, message.ChannelID, content)

		return
	}

	msg, err := s.ChannelMessage(message.ChannelID, reminderMessages[message.ChannelID])
	if err != nil {
		log.Printf(
			"Error getting existing reminder message in %s. Sending a new one. %v\n",
			message.ChannelID,
			err,
		)

		sendMessage(s, message.ChannelID, content)

		return
	}

	if time.Since(message.Timestamp) >= 90*time.Second {
		err = s.ChannelMessageDelete(message.ChannelID, msg.ID)
		if err != nil {
			log.Printf("Error deleting old reminder message in %s: %v\n", message.ChannelID, err)
		}

		sendMessage(s, message.ChannelID, content)
	}
}

func sendMessage(s *discordgo.Session, channelID, content string) {
	msg, err := s.ChannelMessageSend(channelID, content)
	if err != nil {
		log.Printf("Error sending reminder message in %s: %v\n", channelID, err)
		return
	}

	reminderMessages[msg.ChannelID] = msg.ID
}
