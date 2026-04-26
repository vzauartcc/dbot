package helpers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGetMemberName_NoNick(t *testing.T) {
	t.Setenv("LOCAL_DEV_ENVIRONMENT", "true")

	mockMember := &discordgo.Member{
		Nick: "",
		User: &discordgo.User{
			ID:       "1234",
			Username: "Test",
		},
	}

	expected := "Test (1234)"

	newName := GetMemberName(mockMember)

	if newName != expected {
		t.Errorf("Unexpected nickname returned: got %s want %s\n", newName, expected)
	}
}

func TestGetMemberName_Nick(t *testing.T) {
	t.Setenv("LOCAL_DEV_ENVIRONMENT", "true")

	mockMember := &discordgo.Member{
		Nick: "John Doe",
		User: &discordgo.User{
			ID:       "1234",
			Username: "Test",
		},
	}

	expected := "John Doe (1234)"

	newName := GetMemberName(mockMember)

	if newName != expected {
		t.Errorf("Unexpected nickname returned: got %s want %s\n", newName, expected)
	}
}
