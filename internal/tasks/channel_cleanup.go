package tasks

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
)

func (m *Manager) CleanupChannels() {
	log.Println("Starting channel cleanup task")

	for _, cfg := range models.GetConfigs() {
		for channelID, keepMsgID := range cfg.GetCleanupChannels() {
			messages := fetchAllMessages(m.Session, channelID)
			count, errored, total := 0, 0, 0

			for _, msg := range messages {
				if msg.ID != keepMsgID {
					total++

					err := m.Session.ChannelMessageDelete(channelID, msg.ID)
					if err != nil {
						log.Printf(
							"Error deleting message %s in channel %s: %v\n",
							msg.ID,
							channelID,
							err,
						)

						errored++
					} else {
						count++
					}
				}
			}

			log.Printf(
				"Deleted %d messages out of %d total messages (%d failed deletions) in %s\n",
				count,
				total,
				errored,
				channelID,
			)

			time.Sleep(65 * time.Second)
		}
	}
}

func fetchAllMessages(s *discordgo.Session, channelID string) []*discordgo.Message {
	var retval []*discordgo.Message

	beforeID := ""

	for {
		messages, err := s.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			break
		}

		retval = append(retval, messages...)

		if len(messages) < 100 {
			break
		}

		beforeID = messages[len(messages)-1].ID
	}

	return retval
}
