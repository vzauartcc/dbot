package handlers

import (
	"fmt"
	"log"
	"strings"

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
		logCommand(i)

		h(s, i)
	} else {
		log.Printf(
			"Received interaction for a command not registered: /%s (%s)\n",
			commandName,
			helpers.GetMemberName(i.Member),
		)
	}
}

func logCommand(i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Name

	options := i.ApplicationCommandData().Options

	args := make([]string, len(options))
	for i := range options {
		opt := options[i]
		args[i] = fmt.Sprintf("%s: %v", opt.Name, opt.Value)
	}

	optionsString := strings.Join(args, ", ")
	if optionsString == "" {
		optionsString = "none"
	}

	log.Printf(
		"%s ran the command /%s | Options: %s\n",
		helpers.GetMemberName(i.Member),
		name,
		optionsString,
	)
}
