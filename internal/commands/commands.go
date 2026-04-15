package commands

import "github.com/bwmarrin/discordgo"

type HandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

var AllCommands = []*discordgo.ApplicationCommand{
	Ping,
	CID,
	GetRoles,
	Staff,
}

var CommandHandlers = map[string]HandlerFunc{
	"ping":     PingHandler,
	"cid":      CidHandler,
	"getroles": GetRolesHandler,
	"staff":    StaffHandler,
}

func int64Ptr(v int64) *int64 { return &v }
