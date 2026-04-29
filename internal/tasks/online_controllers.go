package tasks

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var waitTimes = []time.Duration{0, 1 * time.Second, 5 * time.Second, 10 * time.Second}

func (m *Manager) UpdateOnlineControllers() {
	online, err := zauapi.GetClient().GetOnlineATC()
	if err != nil {
		log.Printf("Error getting online data: %v\n", err)
		return
	}

	var builder strings.Builder

	for _, controller := range online.Controllers {
		fmt.Fprintf(
			&builder,
			"- %s - %s\n",
			controller.Name,
			secToTime(int(time.Since(controller.LogonTime).Seconds())),
		)
	}

	description := builder.String()
	if strings.TrimSpace(description) == "" {
		description = "No controllers online"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Online Controllers",
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Color:       0x57F287,
	}

	for _, guild := range m.Session.State.Guilds {
		cfg, ok := models.GetConfig(guild.ID)
		if !ok {
			continue
		}

		edit := &discordgo.MessageEdit{
			ID:      cfg.GetOnlineMessage(),
			Channel: cfg.GetOnlineChannel(),
			Embeds:  &[]*discordgo.MessageEmbed{embed},
		}

		for _, delay := range waitTimes {
			time.Sleep(delay)

			_, err := helpers.ChannelMessageEditComplex(m.Session, edit)
			if err == nil {
				return
			}

			// If message is deleted (404), immediately go to send.
			if strings.Contains(err.Error(), "404") {
				break
			}

			log.Printf("Retrying Online Controllers edit due to error: %v\n", err)
		}

		log.Println("Sending new Online Controllers message...")

		sentMsg, err := helpers.ChannelMessageSendEmbed(m.Session, cfg.GetOnlineChannel(), embed)
		if err != nil {
			log.Printf("Error sending new Online Controllers message: %v\n", err)
		} else {
			cfg.SetOnlineMessage(sentMsg.ID, zauapi.GetClient())
		}
	}
}
