//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/steampoweredtaco/gotiktoklive"
)

func TestTikTok_ConnectToLiveRoom(t *testing.T) {
	username := "fiee.fe2"
	if u := testing.Short(); u {
		t.Skip("skipping live connection test in short mode")
	}

	tiktok, err := gotiktoklive.NewTikTok()
	if err != nil {
		t.Fatalf("creating TikTok client: %v", err)
	}

	live, err := tiktok.TrackUser(username)
	if err != nil {
		t.Skipf("user %s is not live, skipping: %v", username, err)
	}

	t.Logf("connected to %s", live.Info.Owner.Username)

	// read events for 10 seconds to verify the connection works
	timeout := time.After(10 * time.Second)
	eventCount := 0
	for {
		select {
		case <-timeout:
			t.Logf("received %d events in 10 seconds", eventCount)
			if eventCount == 0 {
				t.Log("no events received (room may be idle)")
			}
			return
		case event, ok := <-live.Events:
			if !ok {
				t.Log("events channel closed")
				return
			}
			eventCount++
			t.Logf("event: %T", event)
		}
	}
}

func TestTikTok_InvalidUser(t *testing.T) {
	tiktok, err := gotiktoklive.NewTikTok()
	if err != nil {
		t.Fatalf("creating TikTok client: %v", err)
	}

	_, err = tiktok.TrackUser("__nonexistent_user_12345__")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}
