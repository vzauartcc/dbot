package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	helpers "github.com/vzauartcc/dbot/internal/utilities"
)

type HandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

var AllCommands = []*discordgo.ApplicationCommand{
	CID,
	CreateInstructorChannel,
	GetRoles,
	GiveRoles,
	Ping,
	Pirep,
	Staff,
}

var CommandHandlers = map[string]HandlerFunc{
	"create-instructor-channel": CreateInstructorChannelHandler,
	"cid":                       CidHandler,
	"getroles":                  GetRolesHandler,
	"giveroles":                 GetRolesHandler,
	"ping":                      PingHandler,
	"pirep":                     PirepHandler,
	"staff":                     StaffHandler,
}

func int64Ptr(v int64) *int64 {
	p := new(int64)
	*p = v

	return p
}

func sendFollowUpMessage(
	s *discordgo.Session,
	i *discordgo.Interaction,
	command, content string,
) {
	_, err := helpers.FollowupMessageCreate(s, i, true, &discordgo.WebhookParams{
		Content: content,
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		log.Printf(
			"Error sending failure response of /%s for %s: %v\n",
			command,
			helpers.GetMemberName(i.Member),
			err,
		)
	}
}
