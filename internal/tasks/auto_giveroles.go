package tasks

import (
	"log"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

func (m *Manager) AutoGiveRoles() {
	log.Printf("Starting automatic role sync\n")

	cfg, ok := models.GetConfig(helpers.GetMainDiscordServerID())
	if !ok {
		log.Println("Skipping AutoGiveRoles due to no config")
		return
	}

	var syncRole string

	for _, role := range cfg.ManagedRoles {
		if role.LookupKey == "sync" {
			syncRole = role.RoleID
			break
		}
	}

	users, err := zauapi.GetClient().GetUsers()
	if err != nil {
		log.Printf("Error getting users for AutoGiveRoles task: %v\n", err)
		return
	}

	userIDs := make([]string, len(users))
	for i, user := range users {
		userIDs[i] = user.DiscordID
	}

	members := m.FetchGuildMembers(helpers.GetMainDiscordServerID())

	membersByID := make(map[string]*discordgo.Member)
	toRemove := make([]*discordgo.Member, 0)

	for _, member := range members {
		membersByID[member.User.ID] = member

		if syncRole != "" && !slices.Contains(userIDs, member.User.ID) && slices.Contains(member.Roles, syncRole) {
			toRemove = append(toRemove, member)
		}
	}

	for _, user := range users {
		member, ok := membersByID[user.DiscordID]
		if !ok {
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
		if err != nil && !strings.Contains(err.Error(), "HTTP 403 Forbidden") {
			log.Printf(
				"[AutoRoles] Error setting nickname for %s: %v\n",
				helpers.GetMemberName(member),
				err,
			)
		}
	}

	for _, toRemove := range toRemove {
		err := helpers.GuildMemberRoleRemove(m.Session, toRemove.GuildID, toRemove.User.ID, syncRole)
		if err != nil {
			log.Printf("Error removing sync role from %s: %v\n", helpers.GetMemberName(toRemove), err)
		}
	}

	log.Printf("Automatic role sync complete!\n")
}

func (m *Manager) FetchGuildMembers(guildID string) []*discordgo.Member {
	var members []*discordgo.Member

	stop := make(chan struct{})
	nonce := "fetch-members-" + guildID

	removeHandler := helpers.AddHandler(m.Session, func(_ *discordgo.Session, chunk *discordgo.GuildMembersChunk) {
		if chunk.Nonce != nonce {
			return
		}

		members = append(members, chunk.Members...)

		if chunk.ChunkIndex+1 == chunk.ChunkCount {
			close(stop)
		}
	})

	defer removeHandler()

	err := helpers.RequestGuildMembers(m.Session, guildID, "", 0, nonce, false)
	if err != nil {
		log.Printf("Error fetching members: %v\n", err)
		return nil
	}

	<-stop

	return members
}
