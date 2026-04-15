package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleReadyEvent(_ *discordgo.Session, _ *discordgo.Ready) {
	log.Println("Bot is ready!")
}
