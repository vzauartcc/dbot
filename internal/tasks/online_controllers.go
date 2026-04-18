package tasks

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
)

func (m *Manager) UpdateOnlineControllers() {
	log.Println("Running Update Online Controllers task. . . .")

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

	log.Printf("Updating Online Controllers data to:\n%s\n", description)

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

		msg, err := m.Session.ChannelMessage(cfg.GetOnlineChannel(), cfg.GetOnlineMessage())
		if err != nil || len(msg.Embeds) != 1 {
			sentMsg, err := m.Session.ChannelMessageSendEmbed(cfg.GetOnlineChannel(), embed)
			if err != nil {
				log.Printf("Error sending Online Controllers message: %v\n", err)
			} else {
				cfg.SetOnlineMessage(sentMsg.ID, msg.GuildID, zauapi.GetClient())
			}

			return
		}

		edit := &discordgo.MessageEdit{
			ID:      msg.ID,
			Channel: msg.ChannelID,
			Embeds:  &[]*discordgo.MessageEmbed{embed},
		}

		_, err = m.Session.ChannelMessageEditComplex(edit)
		if err != nil {
			log.Printf("Error updating Online Controllers message: %v\n", err)
		}
	}
}
