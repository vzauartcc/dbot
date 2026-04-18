package tasks

import (
	"log"

	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

func (m *Manager) AutoGiveRoles() {
	log.Printf("Starting automatic role sync\n")

	users, err := zauapi.GetClient().GetUsers()
	if err != nil {
		log.Printf("Error getting users for AutoGiveRoles task: %v\n", err)
		return
	}

	for _, user := range users {
		for _, guild := range m.Session.State.Guilds {
			cfg, ok := models.GetConfig(guild.ID)
			if !ok {
				continue
			}

			member, err := m.Session.GuildMember(
				cfg.GetGuildID(),
				user.DiscordID,
			)
			if err != nil {
				continue
			}

			rolesToAdd := helpers.RolesToAdd(cfg, user)

			errs := helpers.ExchangeRoles(m.Session, member, cfg, rolesToAdd, "Scheduled Role Sync")
			for _, err := range errs {
				log.Printf(
					"[AutoRoles] Error syncing role for %s: %v\n",
					helpers.GetMemberName(member),
					err,
				)
			}

			err = helpers.SetNickname(m.Session, member, user)
			if err != nil {
				log.Printf(
					"[AutoRoles] Error setting nickname for %s: %v\n",
					helpers.GetMemberName(member),
					err,
				)
			}
		}
	}

	log.Printf("Automatic role sync complete!\n")
}
