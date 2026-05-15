package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	zauapi "github.com/vzauartcc/dbot/internal/api"
	"github.com/vzauartcc/dbot/internal/api/models"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

type UserData struct {
	ID    string `json:"discord"`
	Token string `json:"token"`
}

var (
	instance        *Instance
	ErrNotConnected = errors.New("redis not connected")
)

type Instance struct {
	client *redis.Client
	ctx    context.Context
}

func ConnectRedis(ctx context.Context) error {
	redisURL := helpers.GetRedisURI()
	if strings.TrimSpace(redisURL) == "" {
		return ErrNotConnected
	}

	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		return err
	}

	redisOpts.ClientName = "dbot"

	instance = &Instance{
		client: redis.NewClient(redisOpts),
		ctx:    ctx,
	}

	return nil
}

func DisconnectRedis() {
	if instance == nil || instance.client == nil {
		return
	}

	instance.client.Close()
}

func GetReminderMessage(channelID string) (string, error) {
	if instance == nil || instance.client == nil {
		return "", ErrNotConnected
	}

	key := "dbot:reminder:" + channelID

	result, err := instance.client.Get(instance.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}

		return "", err
	}

	return result, nil
}

func SetReminderMessage(channelID string, messageID string) error {
	if instance == nil || instance.client == nil {
		return ErrNotConnected
	}

	key := "dbot:reminder:" + channelID

	err := instance.client.Set(instance.ctx, key, messageID, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func StartRedisQueue(s *discordgo.Session) {
	mainGuild := helpers.GetMainDiscordServerID()
	if strings.TrimSpace(mainGuild) == "" {
		log.Println("Redis queue skipped due to no DISCORD_SERVER_ID")
		return
	}

	log.Println("Listening for Discord link events...")

	for {
		if instance == nil || instance.client == nil {
			log.Println("Redis queue aborted due to no Redis connection")
			return
		}

		result, err := instance.client.BRPop(instance.ctx, 0, "dbot:new_discord_user", "dbot:remove_discord_user", "dbot:update_user", "dbot:config_update").Result()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("Redis queue stopped, context closed")
				return
			}

			log.Printf("Error during Redis queue: %v\n", err)

			continue
		}

		queueName := result[0]

		if queueName == "dbot:config_update" {
			log.Println("Received config update event, reloading...")

			cfgs, err := zauapi.GetClient().GetConfigs()
			if err != nil {
				log.Printf("Error getting bot configurations: %v\n", err)
				return
			}

			for _, cfg := range cfgs {
				log.Println("Caching config for guild", cfg.GuildID)

				models.CacheConfig(cfg)
			}

			continue
		}

		cfg, ok := models.GetConfig(mainGuild)
		if !ok {
			continue
		}

		log.Printf("Received Discord link event: %s\n", result[1])

		// User updated by staff or roster-sync.
		if queueName == "dbot:update_user" {
			var user models.User

			err = json.Unmarshal([]byte(result[1]), &user)
			if err != nil {
				log.Printf("Error unmarshaling JSON data for queue: %v\n", err)
				continue
			}

			member, err := helpers.GuildMember(s, mainGuild, user.DiscordID)
			if err != nil {
				log.Printf("[Redis Role Sync] Error getting member for %s: %v\n", user.DiscordID, err)
				continue
			}

			rolesToGive := helpers.RolesToAdd(cfg, user)

			errs := helpers.ExchangeRoles(s, member, cfg, rolesToGive, "Redis Role Sync")
			if len(errs) != 0 {
				log.Printf("Error processing Redis Role Sync for %s: %v\n", user.DiscordID, errs)
			}

			continue
		}

		// Link/unlink event.
		var user UserData

		err = json.Unmarshal([]byte(result[1]), &user)
		if err != nil {
			log.Printf("Error unmarshaling JSON data for queue: %v\n", err)
			continue
		}

		member, err := helpers.GuildMember(s, mainGuild, user.ID)

		if queueName == "dbot:new_discord_user" {
			// User is already a member.
			if err == nil {
				log.Printf(
					"Skipping auto-join for %q: Already in guild.\n",
					strings.ReplaceAll(helpers.GetMemberName(member), "\n", ""),
				)

				continue
			}

			err = helpers.GuildMemberAdd(s, mainGuild, user.ID, &discordgo.GuildMemberAddParams{
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

			continue
		}

		if queueName == "dbot:remove_discord_user" {
			if err != nil {
				// User is not in guild.
				log.Printf("Skipping remove 'sync' role for %s: Not in guild.", user.ID)
				continue
			}

			for _, role := range cfg.GetManagedRoles() {
				if role.LookupKey == "sync" {
					err = helpers.GuildMemberRoleRemove(s, mainGuild, user.ID, role.RoleID)
					if err != nil {
						log.Printf(
							"Error removing 'sync' role from %s: %v\n",
							strings.ReplaceAll(helpers.GetMemberName(member), "\n", ""), err,
						)
					}
				}
			}
		}

		log.Printf("Unknown queue name: %s\n", queueName)
	}
}
