package tasks

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
)

func (m *Manager) UpdateIronMic() {
	log.Println("Running Iron Mic task. . . .")

	data, err := zauapi.GetClient().GetIronMic()
	if err != nil {
		log.Printf("Error update Iron Mic message: %v\n", err)
		return
	}

	embed := buildIronMicMessage(data)

	for _, guild := range m.Session.State.Guilds {
		cfg, ok := zauapi.GetConfig(guild.ID)
		if !ok {
			continue
		}

		msg, err := m.Session.ChannelMessage(cfg.GetIronMicChannel(), cfg.GetIronMicMessage())
		if err != nil || len(msg.Embeds) != 1 {
			sentMsg, err := m.Session.ChannelMessageSendEmbed(cfg.GetIronMicChannel(), embed)
			if err != nil {
				log.Printf("Error sending IronMic message: %v\n", err)
			} else {
				cfg.SetIronMicMessage(sentMsg.ID, msg.GuildID)
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
			log.Printf("Error updating Iron Mic message: %v\n", err)
		}
	}
}

func buildIronMicMessage(data zauapi.IronMicResponse) *discordgo.MessageEmbed {
	var descriptionBuilder strings.Builder
	descriptionBuilder.WriteString("**C1+:**\n")

	for _, controller := range data.Results.Center {
		fmt.Fprintf(&descriptionBuilder,
			"- %s %s - %s\n",
			controller.FirstName,
			controller.LastName,
			secToTime(controller.TotalSeconds),
		)
	}

	descriptionBuilder.WriteString("\n**S3:**\n")

	for _, controller := range data.Results.Approach {
		fmt.Fprintf(&descriptionBuilder,
			"- %s %s - %s\n",
			controller.FirstName,
			controller.LastName,
			secToTime(controller.TotalSeconds),
		)
	}

	descriptionBuilder.WriteString("\n**S2:**\n")

	for _, controller := range data.Results.Tower {
		fmt.Fprintf(&descriptionBuilder,
			"- %s %s - %s\n",
			controller.FirstName,
			controller.LastName,
			secToTime(controller.TotalSeconds),
		)
	}

	descriptionBuilder.WriteString("\n**S1:**\n")

	for _, controller := range data.Results.Ground {
		fmt.Fprintf(&descriptionBuilder,
			"- %s %s - %s\n",
			controller.FirstName,
			controller.LastName,
			secToTime(controller.TotalSeconds),
		)
	}

	log.Printf("Updating Iron Mic body to:\n%s\n", descriptionBuilder.String())

	return &discordgo.MessageEmbed{
		Title:     "Top 3 Quarter Totals for All Ratings",
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf(
				"Calculating %s%d - Updated Hourly.",
				data.Period.Unit,
				data.Period.CurrentPeriod,
			),
		},
		Description: descriptionBuilder.String(),
	}
}

func secToTime(seconds int) string {
	dur := time.Duration(seconds) * time.Second
	h := int(dur.Hours())
	m := int(dur.Minutes()) % 60
	s := int(dur.Seconds()) % 60

	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}
