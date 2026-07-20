package helpers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func SendThinking(s *discordgo.Session, i *discordgo.InteractionCreate, cmd string, ephemeral bool) bool {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := InteractionRespond(s, i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
	if err != nil {
		log.Printf(
			"Error sending thinking response for /%s for %s: %v\n",
			cmd,
			GetMemberName(i.Member),
			err,
		)

		return false
	}

	return true
}
