package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

var Ping = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Check bot latency",
}

func PingHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !helpers.SendThinking(s, i, "ping") {
		return
	}

	latency := helpers.HeartbeatLatency(s)

	_, err := helpers.FollowupMessageCreate(s, i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf(":ping_pong: Pong! My latency is %dms", latency.Milliseconds()),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Printf(
			"Error sending success response for /ping for %s: %v\n",
			helpers.GetMemberName(i.Member),
			err,
		)
	}
}
