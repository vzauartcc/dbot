package commands

import "github.com/bwmarrin/discordgo"

type HandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

var AllCommands = []*discordgo.ApplicationCommand{
	CID,
	GetRoles,
	GiveRoles,
	Ping,
	Pirep,
	Staff,
}

var CommandHandlers = map[string]HandlerFunc{
	"cid":       CidHandler,
	"getroles":  GetRolesHandler,
	"giveroles": GetRolesHandler,
	"ping":      PingHandler,
	"pirep":     PirepHandler,
	"staff":     StaffHandler,
}

func int64Ptr(v int64) *int64 { return &v }
