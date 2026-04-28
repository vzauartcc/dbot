package tasks

import (
	"log"
	"time"

	// Import timezone data.
	_ "time/tzdata"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

type Manager struct {
	Session *discordgo.Session
}

func SetupTasks(s *discordgo.Session) *cron.Cron {
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Printf("Error loading timezone: %v\n", err)
	}

	runner := cron.New(cron.WithLocation(loc))

	manager := &Manager{
		Session: s,
	}

	manager.FetchBotConfigs()

	_, err = runner.AddFunc("*/10 * * * *", manager.AutoGiveRoles)
	if err != nil {
		log.Printf("Error creating AutoGiveRoles task: %v\n", err)
	}

	_, err = runner.AddFunc("1 * * * *", manager.UpdateIronMic)
	if err != nil {
		log.Printf("Error creating UpdateIronMic task: %v\n", err)
	}

	_, err = runner.AddFunc("* * * * *", manager.UpdateOnlineControllers)
	if err != nil {
		log.Printf("Error creating UpdateOnlineControllers task: %v\n", err)
	}

	_, err = runner.AddFunc("12 3 * * *", manager.CleanupChannels)
	if err != nil {
		log.Printf("Error creating ChannelCleanup task: %v\n", err)
	}

	_, err = runner.AddFunc("*/14 * * * *", manager.FetchBotConfigs)
	if err != nil {
		log.Printf("Error creating FetchBotConfigs task: %v\n", err)
	}

	log.Println("Tasks registered!")

	return runner
}
