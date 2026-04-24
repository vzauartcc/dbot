package helpers

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ErrDevEnvironment = errors.New("discord api call skip due to dev environment")

// ApplicationCommands

func ApplicationCommands(s *discordgo.Session, guildID string, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.ApplicationCommands(s.State.User.ID, guildID, options...)
}

func ApplicationCommandCreate(s *discordgo.Session, guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd, options...)
}

func ApplicationCommandDelete(s *discordgo.Session, guildID, cmdID string, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.ApplicationCommandDelete(s.State.User.ID, guildID, cmdID, options...)
}

// Guild Member

func GuildMember(s *discordgo.Session, guildID, userID string, options ...discordgo.RequestOption) (*discordgo.Member, error) {
	// Read-only endpoint; environment check not needed for this API call.
	return s.GuildMember(guildID, userID, options...)
}

func RequestGuildMembers(s *discordgo.Session, guildID, query string, limit int, nonce string, presences bool) error {
	// Read-only endpoint; environment check not needed for this API call.
	return s.RequestGuildMembers(guildID, query, limit, nonce, presences)
}

func GuildMemberAdd(s *discordgo.Session, guildID, userID string, data *discordgo.GuildMemberAddParams, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.GuildMemberAdd(guildID, userID, data, options...)
}

func GuildMemberNickname(s *discordgo.Session, guildID, userID, nickname string, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.GuildMemberNickname(guildID, userID, nickname, options...)
}

func GuildMemberRoleAdd(s *discordgo.Session, guildID, userID, roleID string, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.GuildMemberRoleAdd(guildID, userID, roleID, options...)
}

func GuildMemberRoleRemove(s *discordgo.Session, guildID, userID, roleID string, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.GuildMemberRoleRemove(guildID, userID, roleID, options...)
}

// Guild Roles

func GuildRoles(s *discordgo.Session, guildID string, options ...discordgo.RequestOption) ([]*discordgo.Role, error) {
	// Read-only endpoint; environment check not needed for this API call.
	return s.GuildRoles(guildID, options...)
}

// Handlers

func AddHandler(s *discordgo.Session, handler any) func() {
	if GetIsDevEnvironment() {
		return func() {}
	}

	return s.AddHandler(handler)
}

// Interactions

func InteractionRespond(s *discordgo.Session, interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.InteractionRespond(interaction, resp, options...)
}

func FollowupMessageCreate(s *discordgo.Session, interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.FollowupMessageCreate(interaction, wait, data, options...)
}

// Messages

func ChannelMessage(s *discordgo.Session, channelID, messageID string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	// Read-only endpoint; environment check not needed for this API call.
	return s.ChannelMessage(channelID, messageID, options...)
}

func ChannelMessages(s *discordgo.Session, channelID string, limit int, beforeID, afterID, aroundID string, options ...discordgo.RequestOption) ([]*discordgo.Message, error) {
	// Read-only endpoint; environment check not needed for this API call.
	return s.ChannelMessages(channelID, limit, beforeID, afterID, aroundID, options...)
}

func ChannelMessageEditComplex(s *discordgo.Session, m *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.ChannelMessageEditComplex(m, options...)
}

func ChannelMessageSend(s *discordgo.Session, channelID, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.ChannelMessageSend(channelID, content, options...)
}

func ChannelMessageSendComplex(s *discordgo.Session, channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.ChannelMessageSendComplex(channelID, data, options...)
}

func ChannelMessageSendEmbed(s *discordgo.Session, channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	if GetIsDevEnvironment() {
		return nil, ErrDevEnvironment
	}

	return s.ChannelMessageSendEmbed(channelID, embed, options...)
}

func ChannelMessageDelete(s *discordgo.Session, channelID, messageID string, options ...discordgo.RequestOption) error {
	if GetIsDevEnvironment() {
		return ErrDevEnvironment
	}

	return s.ChannelMessageDelete(channelID, messageID, options...)
}

// Utilities

func HeartbeatLatency(s *discordgo.Session) time.Duration {
	// Read-only endpoint; environment check not needed for this API call.
	return s.HeartbeatLatency()
}
