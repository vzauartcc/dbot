package helpers

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
)

var ratingsToString = []string{
	"SUS",
	"OBS",
	"S1",
	"S2",
	"S3",
	"C1",
	"C2",
	"C3",
	"I1",
	"I2",
	"I3",
	"SUP",
	"ADM",
}

func RolesToAdd(cfg *models.Config, user models.User) []string {
	rolesToGive := make([]string, 0)

	webRoles := user.Roles

	webRoles = append(webRoles, user.CertCodes...)

	if user.IsMember {
		if user.IsVisitor {
			webRoles = append(webRoles, "VIS")
		} else {
			webRoles = append(webRoles, "HOME")
		}
	} else {
		webRoles = append(webRoles, "GUEST")
	}

	rating := ""
	if user.Rating >= 0 && user.Rating <= len(ratingsToString) {
		rating = ratingsToString[user.Rating]
	}

	if strings.ToLower(user.HomeFacility) == "zhq" {
		webRoles = append(webRoles, "zhq")
	}

	webRoles = append(webRoles, "sync")

	for _, roleConfig := range cfg.GetManagedRoles() {
		if slices.Contains(webRoles, roleConfig.LookupKey) {
			rolesToGive = append(rolesToGive, roleConfig.RoleID)
		}

		if rating != "" && roleConfig.LookupKey == rating {
			rolesToGive = append(rolesToGive, roleConfig.RoleID)
		}
	}

	return rolesToGive
}

func calculateRoles(
	cfg *models.Config,
	existingRoles []string,
	rolesToGive []string,
) ([]string, []string) {
	managedRoles := make([]string, 0)

	for _, v := range cfg.GetManagedRoles() {
		if v.RoleID != "" {
			managedRoles = append(managedRoles, v.RoleID)
		}
	}

	rolesToRemove := make([]string, 0)
	rolesToAdd := make([]string, 0)

	for _, role := range existingRoles {
		if slices.Contains(managedRoles, role) && !slices.Contains(rolesToGive, role) {
			rolesToRemove = append(rolesToRemove, role)
		}
	}

	for _, role := range rolesToGive {
		if !slices.Contains(existingRoles, role) {
			rolesToAdd = append(rolesToAdd, role)
		}
	}

	return rolesToAdd, rolesToRemove
}

func ExchangeRoles(
	s *discordgo.Session,
	member *discordgo.Member,
	cfg *models.Config,
	rolesToGive []string,
	reason string,
) []error {
	rolesToAdd, rolesToRemove := calculateRoles(cfg, member.Roles, rolesToGive)

	errors := make([]error, 0)

	if len(rolesToAdd) == 0 && len(rolesToRemove) == 0 {
		return errors
	}

	log.Printf(
		"Role exchange report for %s\n\tAdd: %s\n\tDel: %s\n",
		GetMemberName(member),
		strings.Join(rolesToAdd, ", "),
		strings.Join(rolesToRemove, ", "),
	)

	added := make([]string, 0)

	for _, toAdd := range rolesToAdd {
		if toAdd == "" {
			continue
		}

		err := GuildMemberRoleAdd(s, member.GuildID, member.User.ID, toAdd)
		if err != nil {
			errors = append(errors, fmt.Errorf("error giving role %s: %w", toAdd, err))
		} else {
			added = append(added, toAdd)
		}
	}

	removed := make([]string, 0)

	for _, toRemove := range rolesToRemove {
		if toRemove == "" {
			continue
		}

		err := GuildMemberRoleRemove(s, member.GuildID, member.User.ID, toRemove)
		if err != nil {
			errors = append(errors, fmt.Errorf("error revoking role %s: %w", toRemove, err))
		} else {
			removed = append(removed, toRemove)
		}
	}

	msg := "Role Change Report:\n\n"
	if len(errors) > 0 {
		msg = "Partial Role Change Report (Error giving or taking roles):\n\n"
	}

	allRoles, err := GuildRoles(s, member.GuildID)
	if err != nil {
		log.Printf("Error getting roles for role report: %v\n", err)
		return errors
	}

	rolesAdded := make([]string, len(added))
	for i, role := range added {
		for _, v := range allRoles {
			if v.ID == role {
				rolesAdded[i] = v.Name
			}
		}
	}

	rolesRemoved := make([]string, len(removed))
	for i, role := range removed {
		for _, v := range allRoles {
			if v.ID == role {
				rolesRemoved[i] = v.Name
			}
		}
	}

	_, err = ChannelMessageSend(s,
		"1059182797476077588",
		fmt.Sprintf(
			"%s\n\n%s - Added: %s\n%s - Removed: %s\n\nCause: %s at %s",
			msg,
			GetMemberName(member),
			strings.Join(rolesAdded, ", "),
			GetMemberName(member),
			strings.Join(rolesRemoved, ", "),
			reason,
			time.Now().Format(time.RFC3339),
		),
	)
	if err != nil {
		log.Printf("Error sending role change report message: %v\n", err)
	}

	return errors
}
