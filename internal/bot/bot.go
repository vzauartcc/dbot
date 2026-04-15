package bot

import (
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/commands"
)

func RegisterCommands(s *discordgo.Session) {
	guildID := os.Getenv("DISCORD_SERVER_ID")
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
