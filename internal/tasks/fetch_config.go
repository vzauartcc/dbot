package tasks

import (
	"log"

	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
)

func (m *Manager) FetchBotConfigs() {
	cfgs, err := zauapi.GetClient().GetConfigs()
	if err != nil {
		log.Printf("Error getting bot configurations: %v\n", err)
		return
	}

	for _, cfg := range cfgs {
		models.CacheConfig(cfg)
	}
}
