package bot

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/commands"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

func RegisterCommands(s *discordgo.Session) {
	guildID := helpers.GetMainDiscordServerID()
	if strings.TrimSpace(guildID) == "" {
		log.Println("Skipping command registration due to missing DISCORD_SERVER_ID")
		return
	}

	log.Println("Registering commands...")

	for _, cmd := range commands.AllCommands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			log.Printf("Error registering command /%s: %v\n", cmd.Name, err)
		}

		log.Printf("Registered /%s\n", cmd.Name)
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
