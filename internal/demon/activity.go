package demon

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const maxClientNameBytes = 128

// AddAgentFeeder registers an external host session that will provide its own
// heartbeats. The caller may be an MCP server, Codex, Hermes, or another host.
func (r *Runtime) AddAgentFeeder(client string, pid, parentPID int) (Feeder, error) {
	client = strings.TrimSpace(client)
	if client == "" || len(client) > maxClientNameBytes || strings.ContainsAny(client, "\x00\r\n\t") {
		return Feeder{}, fmt.Errorf("invalid agent client %q", client)
	}
	return r.addFeeder("agent", client, pid, parentPID)
}

func (r *Runtime) addFeeder(kind, client string, pid, parentPID int) (Feeder, error) {
	if kind != "shell" && kind != "agent" {
		return Feeder{}, fmt.Errorf("invalid feeder kind %q", kind)
	}
	if err := r.Ensure(); err != nil {
		return Feeder{}, err
	}
	token, err := newToken()
	if err != nil {
		return Feeder{}, err
	}
	feeder := Feeder{
		Token:     token,
		Kind:      kind,
		Client:    client,
		PID:       pid,
		ParentPID: parentPID,
		Heartbeat: time.Now().UTC(),
	}
	return feeder, atomicJSON(r.feederPath(token), feeder)
}

func validFeederToken(token string) bool {
	if len(token) != 32 || strings.ContainsAny(token, `/\\`) {
		return false
	}
	for _, char := range token {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func (r *Runtime) ReadFeeder(token string) (Feeder, error) {
	if !validFeederToken(token) {
		return Feeder{}, fmt.Errorf("invalid feeder token")
	}
	data, err := os.ReadFile(r.feederPath(token))
	if err != nil {
		return Feeder{}, err
	}
	var feeder Feeder
	if err := json.Unmarshal(data, &feeder); err != nil {
		return Feeder{}, err
	}
	if feeder.Token != token {
		return Feeder{}, fmt.Errorf("feeder token mismatch")
	}
	return feeder, nil
}

// HeartbeatFeeder refreshes one externally managed feeder and returns the
// persisted record so adapters can inspect its current client metadata.
func (r *Runtime) HeartbeatFeeder(token string) (Feeder, error) {
	feeder, err := r.ReadFeeder(token)
	if err != nil {
		return Feeder{}, err
	}
	feeder.Heartbeat = time.Now().UTC()
	if err := atomicJSON(r.feederPath(token), feeder); err != nil {
		return Feeder{}, err
	}
	return feeder, nil
}
