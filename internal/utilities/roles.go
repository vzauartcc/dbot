package helpers

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/vzauartcc/dbot/internal/api/models"
)

var (
	ratingsToString = []string{
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
	ErrInvalidGuildID = errors.New("invalid guild id")
	ErrNoChanges      = errors.New("no roles to give or remove")
	ErrNoData         = errors.New("no roles found on member")
	ErrNoSync         = errors.New("missing sync role")
)

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
	syncRole := "1"

	for _, role := range cfg.GetManagedRoles() {
		if role.LookupKey == "sync" {
			syncRole = role.RoleID
		}
	}

	if !slices.Contains(member.Roles, syncRole) && reason != "/getroles Command" {
		log.Printf("Skipping role sync for %s due to missing SYNC role\n", GetMemberName(member))

		return []error{ErrNoSync}
	}

	rolesToAdd, rolesToRemove := calculateRoles(cfg, member.Roles, rolesToGive)

	if len(rolesToAdd) == 0 && len(rolesToRemove) == 0 {
		return []error{ErrNoChanges}
	}

	errors := make([]error, 0)

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

		err := GuildMemberRoleAdd(s, cfg.GuildID, member.User.ID, toAdd)
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

		err := GuildMemberRoleRemove(s, cfg.GuildID, member.User.ID, toRemove)
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

	allRoles, err := GuildRoles(s, cfg.GuildID)
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

func ProcessStudentAssignment(
	s *discordgo.Session,
	cfg models.Config,
	studentID string,
	instructorCID int,
	added bool,
) {
	instructorRegex := regexp.MustCompile(`cid:(\d+)\s+id:(\d+)$`)

	member, err := GuildMember(s, cfg.GetGuildID(), studentID)
	if err != nil {
		log.Printf("Error getting student: %v\n", err)

		return
	}

	if member == nil {
		log.Println("Student not found")

		return
	}

	allRoles, err := s.GuildRoles(cfg.GetGuildID())
	if err != nil {
		log.Printf("Error getting guild roles: %v\n", err)

		return
	}

	addConfirmed := false
	matched := false

	for _, role := range allRoles {
		if role.Name == "In Training" {
			if added {
				err = GuildMemberRoleAdd(s, cfg.GetGuildID(), studentID, role.ID)
				if err != nil {
					log.Printf(
						"Error adding 'In Training' role to %s: %v\n",
						GetMemberName(member),
						err,
					)
				}
			} else {
				err = GuildMemberRoleRemove(s, cfg.GetGuildID(), studentID, role.ID)
				if err != nil {
					log.Printf(
						"Error removing 'In Training' role from %s: %v\n",
						GetMemberName(member),
						err,
					)
				}
			}

			continue
		}

		// Check to see if the role is the instructor's channel role.
		if !strings.HasSuffix(role.Name, fmt.Sprintf("Student (%d)", instructorCID)) {
			continue
		}

		matched = true

		// Student needs to be added to the instructor's role.
		if added {
			if slices.Contains(member.Roles, role.ID) {
				log.Printf("%s already has %s\n", GetMemberName(member), role.Name)

				continue
			}

			err = GuildMemberRoleAdd(s, cfg.GetGuildID(), studentID, role.ID)
			if err != nil {
				log.Printf("Error adding student role to %s: %v\n", GetMemberName(member), err)

				continue
			}

			log.Printf("Added %s role to %s\n", role.Name, GetMemberName(member))

			addConfirmed = true

			continue
		}

		if !slices.Contains(member.Roles, role.ID) {
			log.Printf("No student role to remove for %s\n", GetMemberName(member))

			continue
		}

		err = GuildMemberRoleRemove(s, cfg.GetGuildID(), studentID, role.ID)
		if err != nil {
			log.Printf("Error removing %s from %s: %v\n", role.Name, GetMemberName(member), err)
		}

		continue
	}

	if !matched {
		log.Printf("Did not find role to modify on student %s\n", GetMemberName(member))
	}

	if !addConfirmed {
		return
	}

	channels, err := s.GuildChannels(cfg.GetGuildID())
	if err != nil {
		log.Printf("Error getting guild channels: %v\n", err)

		return
	}

	for _, channel := range channels {
		matches := instructorRegex.FindStringSubmatch(channel.Topic)

		if len(matches) < 3 || strconv.Itoa(instructorCID) != matches[1] {
			continue
		}

		_, err = ChannelMessageSendComplex(s, channel.ID, &discordgo.MessageSend{
			Content: fmt.Sprintf(
				"<@%s> Welcome to <#%s>! Feel free to post your availability and coordinate training sessions with <@%s> in this channel."+
					"\n\nAs a reminder, you must complete two (2) training sessions per calendar month to remain in the training program. "+
					"The training requirement does not apply to you for this current month."+
					"\n\n**Please post your availability within the next 48 hours.** "+
					"Otherwise your training spot may be given to the next person in line.",
				studentID,
				channel.ID,
				matches[2],
			),
		})
		if err != nil {
			log.Printf("Error sending welcome message for %s: %v\n", GetMemberName(member), err)
		}
	}
}
