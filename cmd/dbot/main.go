package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/bot"
	"github.com/vzauartcc/dbot/internal/queue"
	"github.com/vzauartcc/dbot/internal/tasks"
)

var (
	retryCount int
	maxRetries = 5
	retryMutex sync.Mutex
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	zauapi.Init()

	s, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Println("Invalid bot parameters: ", err)
		return
	}

	bot.RegisterHandlers(s)

	// Special registration for the Ready event to handle retry logic.
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready) {
		retryMutex.Lock()
		defer retryMutex.Unlock()

		retryCount = 0

		log.Println("==========>  Bot is ready!")
	})

	// Special registration for the Resumed event to handle retry logic.
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Resumed) {
		retryMutex.Lock()
		defer retryMutex.Unlock()

		retryCount = 0

		log.Println("==========>  Bot reconnected!")
	})

	// Special registration for the Disconnect event to handle retry logic..
	s.AddHandler(func(_ *discordgo.Session, _ *discordgo.Disconnect) {
		log.Println("==========>  Discord connection lost, waiting for reconnect...")

		retryMutex.Lock()
		defer retryMutex.Unlock()

		retryCount++
		if retryCount > maxRetries {
			log.Printf("Max retries reached, exiting...")
			stop()
		}
	})

	s.Identify.Intents = discordgo.IntentGuilds |
		discordgo.IntentGuildMembers | discordgo.IntentGuildMessages |
		discordgo.IntentMessageContent | discordgo.IntentGuildMessageReactions |
		discordgo.IntentDirectMessages | discordgo.IntentDirectMessageReactions

	s.LogLevel = discordgo.LogWarning

	err = s.Open()
	if err != nil {
		log.Printf("Error opening bot connection: %s\n", err)
		return
	}
	defer s.Close()

	bot.RegisterCommands(s)

	go queue.StartRedisQueue(ctx, s)

	runner := tasks.SetupTasks(s)

	runner.Start()

	log.Printf("Tasks running: %d\n", len(runner.Entries()))

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
