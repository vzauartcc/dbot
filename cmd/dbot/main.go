package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/bot"
	"github.com/vzauartcc/dbot/internal/queue"
	"github.com/vzauartcc/dbot/internal/tasks"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	zauapi.Init()

	s, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Println("Invalid bot parameters: ", err)
		return
	}

	bot.RegisterHandlers(s)

	// Special registration for the Disconnect to call stop().
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Disconnect) {
		log.Println("Bot disconnected, stopping. . . .")
		stop()
	})

	s.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentsMessageContent

	err = s.Open()
	if err != nil {
		log.Printf("Error opening bot connection: %s\n", err)
		return
	}
	defer s.Close()

	bot.RegisterCommands(s, os.Getenv("DISCORD_SERVER_ID"))

	queue.StartRedisQueue(ctx, s)

	runner := tasks.SetupTasks(s)

	runner.Start()

	<-ctx.Done()

	stopCtx := runner.Stop()

	<-stopCtx.Done()

	log.Println("Shutting down...")

	err = bot.UnregisterCommands(s)
	if err != nil {
		log.Printf("Error deleting application commands: %v\n", err)
	}

	log.Println("Application commands deleted!")
}
