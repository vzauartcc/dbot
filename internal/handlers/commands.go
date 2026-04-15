package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/commands"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

func handleApplicationCommands(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	commandName := i.ApplicationCommandData().Name
	if h, ok := commands.CommandHandlers[commandName]; ok {
		log.Printf("%s ran the command /%s\n", helpers.GetMemberName(i.Member), commandName)

		h(s, i)
	} else {
		log.Printf(
			"Received interaction for a command not registered: /%s (%s)\n",
			commandName,
			helpers.GetMemberName(i.Member),
		)
	}
}
