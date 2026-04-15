package helpers

import (
	"slices"
	"testing"

	zauapi "github.com/vzauartcc/dbot/internal/api"
)

func newConfig(roles ...zauapi.ManagedRole) *zauapi.Config {
	return &zauapi.Config{ManagedRoles: roles}
}

func assertContains(t *testing.T, got []string, expected ...string) {
	t.Helper()

	mymap := make(map[string]int)
	for _, v := range got {
		mymap[v]++
	}

	for _, e := range expected {
		if c, ok := mymap[e]; !ok || c == 0 {
			t.Fatalf("expected role %s not found in result %+v", e, got)
		}
	}
}

func assertSliceEqual(t *testing.T, got []string, expected ...string) {
	t.Helper()

	if len(got) != len(expected) {
		t.Fatalf("expected %d items but got %v", len(expected), got)
	}

	for _, e := range expected {
		found := slices.Contains(got, e)
		if !found {
			t.Fatalf("missing %s in %+v", e, got)
		}
	}
}

func TestRolesToAdd_MemberNonVisitor(t *testing.T) {
	cfg := newConfig(
		zauapi.ManagedRole{LookupKey: "HOME", RoleID: "idHome"},
		zauapi.ManagedRole{LookupKey: "SUS", RoleID: "idSus"})

	user := zauapi.User{
		Rating:       0,
		IsMember:     true,
		IsVisitor:    false,
		HomeFacility: "ABC",
	}

	got := RolesToAdd(cfg, user)
	assertContains(t, got, "idHome")
}

func TestRolesToAdd_MemberVis_Cert_Zhq_Rating4(t *testing.T) {
	cfg := newConfig(
		zauapi.ManagedRole{LookupKey: "VIS", RoleID: "idVis"},
		zauapi.ManagedRole{LookupKey: "S3", RoleID: "idS3"},
		zauapi.ManagedRole{LookupKey: "CERT_X", RoleID: "idCertX"},
		zauapi.ManagedRole{LookupKey: "zhq", RoleID: "idZhq"})

	user := zauapi.User{
		Rating:       4,
		IsMember:     true,
		IsVisitor:    true,
		HomeFacility: "zhq",
		CertCodes:    []string{"CERT_X"},
	}

	got := RolesToAdd(cfg, user)
	assertContains(t, got, "idVis", "idS3", "idCertX", "idZhq")
}

func TestRolesToAdd_NonMember_Guest_Rating1(t *testing.T) {
	cfg := newConfig(
		zauapi.ManagedRole{LookupKey: "GUEST", RoleID: "idGuest"},
		zauapi.ManagedRole{LookupKey: "OBS", RoleID: "idObs"})

	user := zauapi.User{
		Rating:       1,
		IsMember:     false,
		HomeFacility: "XYZ",
	}

	got := RolesToAdd(cfg, user)
	assertContains(t, got, "idGuest", "idObs")
}

func TestCalculateRoles_NoChanges(t *testing.T) {
	cfg := newConfig(
		zauapi.ManagedRole{LookupKey: "A", RoleID: "idA"},
		zauapi.ManagedRole{LookupKey: "B", RoleID: "idB"})

	existing := []string{"idA", "idB"}
	expected := []string{"idA", "idB"}

	add, remove := calculateRoles(cfg, existing, expected)
	if len(add) != 0 || len(remove) != 0 {
		t.Fatalf("expected no changes got adds %v removes %v", add, remove)
	}
}

func TestCalculateRoles_AddOnly(t *testing.T) {
	cfg := newConfig(
		zauapi.ManagedRole{LookupKey: "A", RoleID: "idA"},
		zauapi.ManagedRole{LookupKey: "B", RoleID: "idB"},
		zauapi.ManagedRole{LookupKey: "C", RoleID: "idC"})

	existing := []string{"idA"}
	expected := []string{"idA", "idB", "idC"}

	add, remove := calculateRoles(cfg, existing, expected)
	assertSliceEqual(t, add, "idB", "idC")

	if len(remove) != 0 {
		t.Fatalf("expected no removes got %v", remove)
	}
}

func TestCalculateRoles_RemoveOnly(t *testing.T) {
	cfg := newConfig(
		zauapi.ManagedRole{LookupKey: "A", RoleID: "idA"},
		zauapi.ManagedRole{LookupKey: "B", RoleID: "idB"})

	existing := []string{"idA", "idB", "idD"}
	expected := []string{"idA"}

	add, remove := calculateRoles(cfg, existing, expected)
	if len(add) != 0 {
		t.Fatalf("expected no adds got %v", add)
	}

	assertSliceEqual(t, remove, "idB")
}
