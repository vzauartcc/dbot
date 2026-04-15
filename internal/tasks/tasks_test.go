package tasks

import "testing"

func TestSecConverter(t *testing.T) {
	ninety := secToTime(5400)
	if ninety != "1h 30m 0s" {
		t.Error("Unexpected output", ninety)
	}

	ninety3 := secToTime(5403)
	if ninety3 != "1h 30m 3s" {
		t.Error("Unexpected output", ninety)
	}
}
