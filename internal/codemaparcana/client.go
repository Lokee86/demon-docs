package codemaparcana

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const protocolID = "arcana.query.v1"

type protocolClient struct {
	mu     sync.Mutex
	stdin  io.WriteCloser
	stdout *bufio.Reader
	cmd    *exec.Cmd
	stderr lockedBuffer
	nextID uint64
	closed bool
}

type protocolResponse struct {
	Protocol string          `json:"protocol"`
	ID       json.RawMessage `json:"id"`
	OK       bool            `json:"ok"`
	Result   json.RawMessage `json:"result"`
	Error    *protocolError  `json:"error"`
}

type protocolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func startProtocolClient(ctx context.Context, command, snapshotDirectory string) (*protocolClient, error) {
	cmd := exec.CommandContext(ctx, command, "protocol", "--snapshot", snapshotDirectory)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	client := &protocolClient{stdin: stdin, stdout: bufio.NewReader(stdout), cmd: cmd}
	cmd.Stderr = &client.stderr
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, err
	}
	var capabilities struct {
		Protocol   string   `json:"protocol"`
		Version    int      `json:"version"`
		Operations []string `json:"operations"`
	}
	if err := client.query(ctx, map[string]any{"op": "capabilities"}, &capabilities); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("query Arcana capabilities: %w", err)
	}
	if capabilities.Protocol != protocolID || capabilities.Version != 1 ||
		!containsString(capabilities.Operations, "resolve_file") ||
		!containsString(capabilities.Operations, "resolve_symbol") ||
		!containsString(capabilities.Operations, "list_nodes") ||
		!containsString(capabilities.Operations, "neighbors") {
		_ = client.Close()
		return nil, fmt.Errorf("Arcana protocol does not provide required codemap operations")
	}
	return client, nil
}

func (client *protocolClient) query(ctx context.Context, request map[string]any, target any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closed {
		return fmt.Errorf("Arcana protocol client is closed")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	client.nextID++
	request["id"] = client.nextID
	encoded, err := json.Marshal(request)
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	if _, err := client.stdin.Write(encoded); err != nil {
		return client.transportError("write request", err)
	}
	line, err := client.stdout.ReadBytes('\n')
	if err != nil {
		return client.transportError("read response", err)
	}
	var response protocolResponse
	if err := json.Unmarshal(bytes.TrimSpace(line), &response); err != nil {
		return fmt.Errorf("decode Arcana response: %w", err)
	}
	if response.Protocol != protocolID {
		return fmt.Errorf("unexpected Arcana protocol %q", response.Protocol)
	}
	var responseID uint64
	if err := json.Unmarshal(response.ID, &responseID); err != nil || responseID != client.nextID {
		return fmt.Errorf("Arcana response ID does not match request")
	}
	if !response.OK {
		if response.Error == nil {
			return fmt.Errorf("Arcana query failed without an error payload")
		}
		return fmt.Errorf("Arcana query failed [%s]: %s", response.Error.Code, response.Error.Message)
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(response.Result, target); err != nil {
		return fmt.Errorf("decode Arcana result: %w", err)
	}
	return nil
}

func (client *protocolClient) Close() error {
	client.mu.Lock()
	if client.closed {
		client.mu.Unlock()
		return nil
	}
	client.closed = true
	err := client.stdin.Close()
	client.mu.Unlock()
	waitErr := client.cmd.Wait()
	if err != nil {
		return err
	}
	if waitErr != nil && client.cmd.ProcessState != nil && client.cmd.ProcessState.ExitCode() != 0 {
		return client.transportError("wait for Arcana protocol", waitErr)
	}
	return nil
}

func (client *protocolClient) transportError(operation string, err error) error {
	message := strings.TrimSpace(client.stderr.String())
	if message == "" {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return fmt.Errorf("%s: %w: %s", operation, err, message)
}

func findArcanaCommand() (string, bool) {
	if configured := strings.TrimSpace(os.Getenv("DDOCS_ARCANA_COMMAND")); configured != "" {
		return configured, true
	}
	if executable, err := os.Executable(); err == nil {
		name := "arcana"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		adjacent := filepath.Join(filepath.Dir(executable), name)
		if info, statErr := os.Stat(adjacent); statErr == nil && !info.IsDir() {
			return adjacent, true
		}
	}
	command, err := exec.LookPath("arcana")
	return command, err == nil
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
