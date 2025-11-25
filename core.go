package amelecore

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/vmihailenco/msgpack/v5"
)

var conn net.Conn
var decoder *msgpack.Decoder
var storedContext map[string]any

func Accept() (map[string]any, error) {
	if os.Getenv("COMMUNICATION_PROTOCOL") == "tcp" {
		port := os.Getenv("AMELE_TCP_PORT")
		var err error
		conn, err = net.Dial("tcp", "127.0.0.1:"+port)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to orchestrator: %v", err)
		}

		decoder = msgpack.NewDecoder(conn)
		var envelope map[string]any
		if err := decoder.Decode(&envelope); err != nil {
			return nil, fmt.Errorf("failed to decode envelope: %v", err)
		}

		// Extract context and inputs from envelope
		if ctx, ok := envelope["context"].(map[string]any); ok {
			storedContext = ctx
		} else {
			storedContext = map[string]any{}
		}

		if inputs, ok := envelope["inputs"].(map[string]any); ok {
			return inputs, nil
		}
		return map[string]any{}, nil
	}

	// For shmem mode, read from inbox file
	shmemPath := os.Getenv("AMELE_INBOX_FILE")
	if shmemPath == "" {
		return map[string]any{}, nil
	}

	data, err := os.ReadFile(shmemPath)
	if err != nil {
		return nil, fmt.Errorf("error reading inbox file: %v", err)
	}

	var envelope map[string]any
	if err := msgpack.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("error unmarshaling envelope: %v", err)
	}

	// Extract context and inputs from envelope
	if ctx, ok := envelope["context"].(map[string]any); ok {
		storedContext = ctx
	} else {
		storedContext = map[string]any{}
	}

	if inputs, ok := envelope["inputs"].(map[string]any); ok {
		return inputs, nil
	}
	return map[string]any{}, nil
}

func Context() map[string]any {
	return storedContext
}

func CallFunction(name string, inputs map[string]any) (map[string]any, error) {
	if os.Getenv("COMMUNICATION_PROTOCOL") == "tcp" {
		if conn == nil {
			return nil, fmt.Errorf("connection not initialized")
		}

		bytes := make([]byte, 4)
		rand.Read(bytes)
		reqID := hex.EncodeToString(bytes)

		req := map[string]any{
			"type":     "call",
			"function": name,
			"inputs":   inputs,
			"id":       reqID,
		}

		if err := msgpack.NewEncoder(conn).Encode(req); err != nil {
			return nil, fmt.Errorf("failed to send call request: %v", err)
		}

		var resp map[string]any
		if err := decoder.Decode(&resp); err != nil {
			return nil, fmt.Errorf("failed to receive call response: %v", err)
		}

		if resp["type"] == "call_result" && resp["id"] == reqID {
			if errMsg, ok := resp["error"].(string); ok {
				return nil, fmt.Errorf("%s", errMsg)
			}
			if result, ok := resp["result"].(map[string]any); ok {
				return result, nil
			}
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected response: %v", resp)
	}
	return nil, fmt.Errorf("callFunction is not supported in shmem mode")
}

func Respond(context map[string]any) error {
	if os.Getenv("COMMUNICATION_PROTOCOL") == "tcp" {
		if conn == nil {
			return fmt.Errorf("connection not initialized")
		}
		msg := map[string]any{
			"type":    "respond",
			"context": context,
		}
		return msgpack.NewEncoder(conn).Encode(msg)
	}

	shmemPath := os.Getenv("AMELE_OUTBOX_FILE")
	if shmemPath == "" {
		return fmt.Errorf("AMELE_OUTBOX_FILE not set")
	}

	data, err := msgpack.Marshal(context)
	if err != nil {
		return fmt.Errorf("error marshaling context: %v", err)
	}

	if err := os.WriteFile(shmemPath, data, 0600); err != nil {
		return fmt.Errorf("error writing outbox file: %v", err)
	}

	return nil
}
