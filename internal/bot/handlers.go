package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/handlers"
)

func RegisterHandlers(s *discordgo.Session) {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handlers.HandleInteractions(s, i)
	})

	s.AddHandler(func(s *discordgo.Session, message *discordgo.MessageCreate) {
		handlers.HandleMessage(s, message)
	})

	s.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		handlers.HandleReadyEvent(s, ready)
	})
}
