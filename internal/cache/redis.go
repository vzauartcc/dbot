package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
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
	redisOpts.MaintNotificationsConfig = &maintnotifications.Config{
		Mode: maintnotifications.ModeDisabled,
	}

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
	if strings.TrimSpace(helpers.GetMainDiscordServerID()) == "" {
		log.Println("Redis queue skipped due to no DISCORD_SERVER_ID")

		return
	}

	log.Println("Listening for Discord link events...")

	for {
		if instance == nil || instance.client == nil {
			log.Println("Redis queue aborted due to no Redis connection")

			return
		}

		result, err := instance.client.BRPop(instance.ctx, 0, "dbot:new_discord_user", "dbot:remove_discord_user", "dbot:update_user", "dbot:config_update", "dbot:student_assignment").
			Result()
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
			handleConfigUpdate()

			continue
		}

		cfg, ok := models.GetConfig(helpers.GetMainDiscordServerID())
		if !ok {
			log.Printf("Skipping %s due to missing config.\n", queueName)

			continue
		}

		switch queueName {
		case "dbot:student_assignment":
			handleStudentAssignment(s, result[1], cfg)
		case "dbot:update_user":
			handleUserUpdate(s, result[1], cfg)
		case "dbot:new_discord_user":
			handleUserAdd(s, result[1], cfg)
		case "dbot:remove_discord_user":
			handleUserRemove(s, result[1], cfg)
		default:
			log.Printf("Unknown queue event: %s\n", queueName)
		}
	}
}

func handleConfigUpdate() {
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
}

func handleStudentAssignment(s *discordgo.Session, data string, cfg *models.Config) {
	var assignment models.StudentAssignment

	err := json.Unmarshal([]byte(data), &assignment)
	if err != nil {
		log.Printf("Error unmarshaling JSON data for student_assignment queue: %v\n", err)

		return
	}

	if assignment.StudentID != "" && assignment.InstructorCID != 0 {
		helpers.ProcessStudentAssignment(
			s,
			*cfg,
			assignment.StudentID,
			assignment.InstructorCID,
			assignment.Added,
		)
	}
}

func handleUserUpdate(s *discordgo.Session, data string, cfg *models.Config) {
	var user models.User

	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		log.Printf("Error unmarshaling JSON data for queue: %v\n", err)

		return
	}

	if strings.TrimSpace(user.DiscordID) == "" {
		return
	}

	log.Printf(
		"Processing user update for %s %s (%s)\n",
		user.FirstName,
		user.LastName,
		user.DiscordID,
	)

	member, err := helpers.GuildMember(s, helpers.GetMainDiscordServerID(), user.DiscordID)
	if err != nil {
		log.Printf(
			"[Redis Role Sync] Error getting member for %s: %v\n",
			user.DiscordID,
			err,
		)

		return
	}

	rolesToGive := helpers.RolesToAdd(cfg, user)

	errs := helpers.ExchangeRoles(s, member, cfg, rolesToGive, "Redis Role Sync")
	if len(errs) != 0 {
		log.Printf("Error processing Redis Role Sync for %s: %v\n", user.DiscordID, errs)
	}
}

func handleUserAdd(s *discordgo.Session, data string, cfg *models.Config) {
	var syncRole string

	for _, role := range cfg.GetManagedRoles() {
		if role.LookupKey == "sync" {
			syncRole = role.RoleID
		}
	}

	// Link/unlink events.
	var user UserData

	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		log.Printf("Error unmarshaling JSON data for queue: %v\n", err)

		return
	}

	if strings.TrimSpace(user.ID) == "" {
		log.Println("Skipping new_discord_user due to no Discord ID")

		return
	}

	member, err := helpers.GuildMember(s, helpers.GetMainDiscordServerID(), user.ID)

	// User is already a member.
	if err == nil {
		log.Printf(
			"Skipping auto-join for %q: Already in guild.\n",
			strings.ReplaceAll(helpers.GetMemberName(member), "\n", ""),
		)

		err = helpers.GuildMemberRoleAdd(s, cfg.GuildID, user.ID, syncRole, nil)
		if err != nil {
			log.Printf(
				"Error adding sync role during link event to %s",
				helpers.GetMemberName(member),
			)
		}

		return
	}

	err = helpers.GuildMemberAdd(
		s,
		helpers.GetMainDiscordServerID(),
		user.ID,
		&discordgo.GuildMemberAddParams{
			AccessToken: user.Token,
			Nick:        "",
			Mute:        false,
			Deaf:        false,
			Roles:       []string{syncRole},
		},
	)
	if err != nil {
		log.Printf("Error auto-joining %s to guild: %v\n", user.ID, err)

		return
	}

	log.Printf("Joined %s to the guild!\n", user.ID)
}

func handleUserRemove(s *discordgo.Session, data string, cfg *models.Config) {
	var user UserData

	err := json.Unmarshal([]byte(data), &user)
	if err != nil {
		log.Printf("Error unmarshaling JSON data for queue: %v\n", err)

		return
	}

	if strings.TrimSpace(user.ID) == "" {
		log.Println("Skipping remove_discord_user due to no Discord ID")

		return
	}

	member, err := helpers.GuildMember(s, helpers.GetMainDiscordServerID(), user.ID)
	if err != nil {
		// User is not in guild.
		log.Printf("Skipping remove 'sync' role for %s: Not in guild.", user.ID)

		return
	}

	for _, role := range cfg.GetManagedRoles() {
		if role.LookupKey == "sync" {
			err = helpers.GuildMemberRoleRemove(
				s,
				helpers.GetMainDiscordServerID(),
				user.ID,
				role.RoleID,
			)
			if err != nil {
				log.Printf(
					"Error removing 'sync' role from %s: %v\n",
					strings.ReplaceAll(helpers.GetMemberName(member), "\n", ""), err,
				)
			}
		}
	}
}
