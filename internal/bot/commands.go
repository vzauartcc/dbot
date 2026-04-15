package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/commands"
)

func HandleApplicationCommands(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) {
	commandName := i.ApplicationCommandData().Name
	if h, ok := commands.CommandHandlers[commandName]; ok {
		h(s, i)
	}
}

func UnregisterCommands(s *discordgo.Session) error {
	commands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		return err
	}

	for _, v := range commands {
		_ = s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
	}

	for _, guild := range s.State.Guilds {
		cmds, err := s.ApplicationCommands(s.State.User.ID, guild.ID)
		if err != nil {
			continue
		}

		for _, v := range cmds {
			_ = s.ApplicationCommandDelete(s.State.User.ID, guild.ID, v.ID)
		}
	}

	return nil
}
