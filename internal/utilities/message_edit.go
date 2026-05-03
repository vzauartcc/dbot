package helpers

import (
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var waitTimes = []time.Duration{0, 1 * time.Second, 5 * time.Second, 10 * time.Second}

func TryMessageEdit(s *discordgo.Session, edit *discordgo.MessageEdit, msg string) error {
	var err error

	for _, delay := range waitTimes {
		time.Sleep(delay)

		_, err = ChannelMessageEditComplex(s, edit)
		if err == nil {
			return nil
		}

		// If message is deleted (404), immediately go to send.
		if strings.Contains(err.Error(), "404") {
			break
		}

		log.Printf("Retrying %s message edit due to error: %v\n", msg, err)
	}

	return err
}
