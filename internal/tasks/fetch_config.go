package tasks

import (
	"log"

	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
)

func (m *Manager) FetchBotConfigs() {
	log.Println("Fetching all configs...")

	cfgs, err := zauapi.GetClient().GetConfigs()
	if err != nil {
		log.Printf("Error getting bot configurations: %v\n", err)
		return
	}

	for _, cfg := range cfgs {
		log.Println("Caching config for guild", cfg.GuildID)

		models.CacheConfig(cfg)
	}
}
