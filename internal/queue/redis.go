package queue

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

type UserData struct {
	ID    string `json:"discord"`
	Token string `json:"token"`
}

func StartRedisQueue(ctx context.Context, s *discordgo.Session) {
	mainGuild := os.Getenv("DISCORD_SERVER_ID")
	if strings.TrimSpace(mainGuild) == "" {
		log.Println("Redis queue skipped due to no DISCORD_SERVER_ID")
		return
	}

	redisURL := os.Getenv("REDIS_URI")
	if strings.TrimSpace(redisURL) == "" {
		log.Println("Redis queue skipped due to no REDIS_URI")
	}

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Redis queue skipped due to Redis URL parsing error: %v\n", err)
		return
	}

	rdb := redis.NewClient(redisOpts)

	for {
		result, err := rdb.BRPop(ctx, 0, "new_discord_user", "remove_discord_user").Result()
		if err != nil {
			log.Printf("Error during Redis queue: %v\n", err)
			continue
		}

		var user UserData

		err = json.Unmarshal([]byte(result[1]), &user)
		if err != nil {
			log.Printf("Error unmarshalling JSON data for queue: %v\n", err)
			continue
		}

		queueName := result[0]
		member, err := s.GuildMember(mainGuild, user.ID)

		if queueName == "new_discord_user" {
			// User is already a member.
			if err == nil {
				log.Printf(
					"Skipping auto-join for %q: Already in guild.\n",
					strings.ReplaceAll(helpers.GetMemberName(member), "\n", ""),
				)

				continue
			}

			err = s.GuildMemberAdd(mainGuild, user.ID, &discordgo.GuildMemberAddParams{
				AccessToken: user.Token,
				Nick:        "",
				Mute:        false,
				Deaf:        false,
				Roles:       nil,
			})
			if err != nil {
				log.Printf("Error auto-joining %s to guild: %v\n", user.ID, err)
			} else {
				log.Printf("Joined %s to the guild!\n", user.ID)
			}
		} else if queueName == "remove_discord_user" {
			if err != nil {
				// User is not in guild.
				log.Printf("Skipping remove 'sync' role for %s: Not in guild.", user.ID)
				continue
			}

			cfg, ok := zauapi.GetConfig(mainGuild)
			if !ok {
				continue
			}

			for _, role := range cfg.GetManagedRoles() {
				if role.LookupKey == "sync" {
					err = s.GuildMemberRoleRemove(mainGuild, user.ID, role.RoleID)
					if err != nil {
						log.Printf(
							"Error removing 'sync' role from %s: %v\n",
							strings.ReplaceAll(helpers.GetMemberName(member), "\n", ""), err,
						)
					}
				}
			}
		}
	}
}
