package models

import (
	"log"
	"maps"
	"slices"
	"sync"
)

type Config struct {
	GuildID           string            `json:"id"`
	RepostChannels    map[string]string `json:"repostChannels"`
	CleanupChannels   map[string]string `json:"cleanupChannels"`
	IronMicConfig     UpdateableMessage `json:"ironMic"`
	OnlineControllers UpdateableMessage `json:"onlineControllers"`
	ManagedRoles      []ManagedRole     `json:"managedRoles"`
	ReminderChannels  map[string]string `json:"reminderChannels"`
}

type UpdateableMessage struct {
	ChannelID string `json:"channelId"`
	MessageID string `json:"messageId"`
}

type ManagedRole struct {
	LookupKey string `json:"key"`
	RoleID    string `json:"roleId"`
}

// ConfigUpdater is an interface to allow API interaction,
// its functions should match the API.
type ConfigUpdater interface {
	UpdateConfig(guildID string, config *Config) (*Config, error)
}

type ConfigUpdate struct {
	IronMicMessageID           string `json:"ironMic"`
	OnlineControllersMessageID string `json:"onlineControllers"`
}

var configs map[string]*Config
var mutex sync.RWMutex

func GetConfig(guildID string) (*Config, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	if cfg, ok := configs[guildID]; ok {
		return cfg, true
	}

	return nil, false
}

func GetConfigs() []*Config {
	mutex.RLock()
	defer mutex.RUnlock()

	return slices.Collect(maps.Values(configs))
}

func CacheConfig(config Config) {
	mutex.Lock()
	defer mutex.Unlock()

	if configs == nil {
		configs = make(map[string]*Config)
	}

	configs[config.GuildID] = &config
}

func (c *Config) GetGuildID() string {
	return c.GuildID
}

func (c *Config) GetCleanupChannels() map[string]string {
	return c.CleanupChannels
}

func (c *Config) GetRepostChannels() map[string]string {
	return c.RepostChannels
}

func (c *Config) GetReminderChannels() map[string]string {
	return c.ReminderChannels
}

func (c *Config) GetIronMicChannel() string {
	return c.IronMicConfig.ChannelID
}

func (c *Config) GetIronMicMessage() string {
	return c.IronMicConfig.MessageID
}

func (c *Config) SetIronMicMessage(messageID string, api ConfigUpdater) {
	mutex.Lock()
	defer mutex.Unlock()

	old := c.IronMicConfig.MessageID
	c.IronMicConfig.MessageID = messageID

	c.updateConfig(api)

	log.Printf("Updated Iron Mic message from %s to %s\n", old, messageID)
}

func (c *Config) GetManagedRoles() []ManagedRole {
	return c.ManagedRoles
}

func (c *Config) GetOnlineChannel() string {
	return c.OnlineControllers.ChannelID
}

func (c *Config) GetOnlineMessage() string {
	return c.OnlineControllers.MessageID
}

func (c *Config) SetOnlineMessage(messageID string, api ConfigUpdater) {
	mutex.Lock()
	defer mutex.Unlock()

	old := c.OnlineControllers.MessageID
	c.OnlineControllers.MessageID = messageID

	c.updateConfig(api)

	log.Printf("Updated Online Controllers message from %s to %s\n", old, messageID)
}

func (c *Config) updateConfig(service ConfigUpdater) {
	_, err := service.UpdateConfig(c.GuildID, c)
	if err != nil {
		log.Printf("Error updating config: %v\n", err)
		return
	}

	log.Println("Successfully updated config!")
}
