package tasks

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

type Manager struct {
	Session *discordgo.Session
}

func SetupTasks(s *discordgo.Session) *cron.Cron {
	runner := cron.New(cron.WithLocation(time.FixedZone("America/Chicago", 0)))

	manager := &Manager{
		Session: s,
	}

	manager.FetchBotConfigs()

	_, err := runner.AddFunc("*/10 * * * *", manager.AutoGiveRoles)
	if err != nil {
		log.Printf("Error creating AutoGiveRoles task: %v\n", err)
	}

	_, err = runner.AddFunc("0 * * * *", manager.UpdateIronMic)
	if err != nil {
		log.Printf("Error creating UpdateIronMic task: %v\n", err)
	}

	_, err = runner.AddFunc("* * * * *", manager.UpdateOnlineControllers)
	if err != nil {
		log.Printf("Error creating UpdateOnlineControllers task: %v\n", err)
	}

	_, err = runner.AddFunc("0 3 * * *", manager.CleanupChannels)
	if err != nil {
		log.Printf("Error creating ChannelCleanup task: %v\n", err)
	}

	_, err = runner.AddFunc("0 0 * * *", manager.FetchBotConfigs)
	if err != nil {
		log.Printf("Error creating FetchBotConfigs task: %v\n", err)
	}

	log.Println("Tasks registered!")

	return runner
}
