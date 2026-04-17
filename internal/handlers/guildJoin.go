package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

func HandleGuildMemberAdd(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
	cfg, ok := models.GetConfig(event.GuildID)
	if !ok {
		return
	}

	user, err := zauapi.GetClient().GetUserByID(event.User.ID)
	if err != nil {
		log.Printf(
			"Error handling GuildMemberAdd event for %s: %v\n",
			helpers.GetMemberName(event.Member),
			err,
		)

		return
	}

	rolesToAdd := helpers.RolesToAdd(cfg, user)

	errs := helpers.ExchangeRoles(s, event.Member, cfg, rolesToAdd, "GuildMemberAdd Event")
	for _, err := range errs {
		log.Printf(
			"[GuildMemberAdd] Error syncing role for %s: %v\n",
			helpers.GetMemberName(event.Member),
			err,
		)
	}

	err = helpers.SetNickname(s, event.Member, user)
	if err != nil {
		log.Printf(
			"[GuildMemberAdd] Error setting nickname for %s: %v\n",
			helpers.GetMemberName(event.Member),
			err,
		)
	}
}
