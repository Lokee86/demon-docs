package demon

import (
	"os"
	"testing"
	"time"
)

func TestAgentFeederLifecycle(t *testing.T) {
	runtime := testRuntime(t)
	feeder, err := runtime.AddAgentFeeder("mcp", 10, 20)
	if err != nil {
		t.Fatal(err)
	}
	if feeder.Kind != "agent" || feeder.Client != "mcp" || feeder.PID != 10 || feeder.ParentPID != 20 {
		t.Fatalf("unexpected feeder: %+v", feeder)
	}

	aged := feeder
	aged.Heartbeat = time.Now().Add(-time.Minute)
	if err := atomicJSON(runtime.feederPath(feeder.Token), aged); err != nil {
		t.Fatal(err)
	}
	updated, err := runtime.HeartbeatFeeder(feeder.Token)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Client != "mcp" || !updated.Heartbeat.After(aged.Heartbeat) {
		t.Fatalf("heartbeat lost feeder metadata: %+v", updated)
	}
	persisted, err := runtime.ReadFeeder(feeder.Token)
	if err != nil || persisted.Heartbeat != updated.Heartbeat {
		t.Fatalf("persisted=%+v err=%v", persisted, err)
	}

	if err := runtime.RemoveFeeder(feeder.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.ReadFeeder(feeder.Token); !os.IsNotExist(err) {
		t.Fatalf("released feeder remained: %v", err)
	}
}

func TestAgentFeederRejectsInvalidClientAndToken(t *testing.T) {
	runtime := testRuntime(t)
	for _, client := range []string{"", "   ", "mcp\nother"} {
		if _, err := runtime.AddAgentFeeder(client, 1, 2); err == nil {
			t.Fatalf("accepted invalid client %q", client)
		}
	}
	for _, token := range []string{"", "../owner", "not-a-token"} {
		if _, err := runtime.HeartbeatFeeder(token); err == nil {
			t.Fatalf("accepted invalid token %q", token)
		}
		if err := runtime.RemoveFeeder(token); err == nil {
			t.Fatalf("released invalid token %q", token)
		}
	}
}
