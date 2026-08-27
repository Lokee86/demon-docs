package codemaparcana

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
)

func TestProtocolClientFramesAndMatchesJSONLResponse(t *testing.T) {
	requestReader, requestWriter := io.Pipe()
	responseReader, responseWriter := io.Pipe()
	client := &protocolClient{stdin: requestWriter, stdout: bufio.NewReader(responseReader)}
	done := make(chan error, 1)
	go func() {
		defer requestReader.Close()
		defer responseWriter.Close()
		line, err := bufio.NewReader(requestReader).ReadBytes('\n')
		if err != nil {
			done <- err
			return
		}
		var request map[string]any
		if err := json.Unmarshal(line, &request); err != nil {
			done <- err
			return
		}
		response := map[string]any{
			"protocol": protocolID,
			"id":       request["id"],
			"ok":       true,
			"result":   map[string]any{"count": 1},
		}
		encoded, err := json.Marshal(response)
		if err == nil {
			_, err = responseWriter.Write(append(encoded, '\n'))
		}
		done <- err
	}()

	var result struct {
		Count int `json:"count"`
	}
	if err := client.query(context.Background(), map[string]any{"op": "resolve_file", "path": "main.go"}, &result); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if result.Count != 1 {
		t.Fatalf("result = %#v", result)
	}
}
