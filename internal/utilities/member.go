package helpers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func GetMemberName(m *discordgo.Member) string {
	if strings.TrimSpace(m.Nick) != "" {
		return fmt.Sprintf("%s (%s)", m.Nick, m.User.ID)
	}

	return fmt.Sprintf("%s (%s)", m.User.Username, m.User.ID)
}
