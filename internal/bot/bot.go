package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/commands"
)

func RegisterCommands(s *discordgo.Session, guildID string) {
	log.Println("Registering commands...")

	for _, cmd := range commands.AllCommands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			log.Printf("Error registering command /%s: %v\n", cmd.Name, err)
		}

		log.Printf("Registered /%s\n", cmd.Name)
	}
}
