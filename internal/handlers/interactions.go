package handlers

import (
	"github.com/bwmarrin/discordgo"
)

func HandleInteractions(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	if i.Type == discordgo.InteractionApplicationCommand {
		handleApplicationCommands(s, i)
	}
}
