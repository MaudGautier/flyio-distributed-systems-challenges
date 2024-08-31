package main

import (
	"context"
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"os"
)

func main() {

	node := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(node)
	ctx := context.Background()
	lastWrittenValue := 0 // TODO voir quoi mettre

	node.Handle("init", func(msg maelstrom.Message) error {
		// Initialize global counter value
		_, err := kv.Read(ctx, "global-counter")
		if err != nil {
			kv.Write(ctx, "global-counter", 0)
		}

		return nil
	})

	/*
		// Add - input message body
		{
			"type": "add",
			"delta": 123
		}

		// Add - response
		{
			"type": "add_ok",
		}
	*/
	node.Handle("add", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		newCounterValue, _ := updateValue(ctx, kv, int(body["delta"].(float64)))

		lastWrittenValue = newCounterValue

		// Create return body
		returnBody := make(map[string]interface{})
		returnBody["type"] = "add_ok"

		// Echo the original message back with the updated message type.
		return node.Reply(msg, returnBody)
	})

	/*
		// Read - input message body
		{
			"type": "read"
		}

		// Read - response
		{
			"type": "read_ok",
			"value": 1234
		}
	*/

	node.Handle("read", func(msg maelstrom.Message) error {
		// Get counter value from other nodes
		mostUpToDateValue := 0
		body := make(map[string]interface{})
		body["type"] = "last_written"
		for _, otherNode := range node.NodeIDs() {
			result, _ := node.SyncRPC(ctx, otherNode, body)
			// Unmarshal the message body as a loosely-typed map.
			var resultBody map[string]any
			if err := json.Unmarshal(result.Body, &resultBody); err != nil {
				return err
			}
			lastWrittenByNode := int(resultBody["last_written"].(float64))

			if mostUpToDateValue < lastWrittenByNode {
				mostUpToDateValue = lastWrittenByNode
			}
		}

		// Create return body
		returnBody := make(map[string]interface{})
		returnBody["type"] = "read_ok"
		returnBody["value"] = mostUpToDateValue

		// Echo the original message back with the updated message type.
		return node.Reply(msg, returnBody)
	})

	node.Handle("last_written", func(msg maelstrom.Message) error {
		// Create return body
		returnBody := make(map[string]interface{})
		returnBody["type"] = "last_written_ok"
		returnBody["last_written"] = lastWrittenValue

		return node.Reply(msg, returnBody)
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}

}

func updateValue(ctx context.Context, kv *maelstrom.KV, delta int) (int, error) {
	// Get counter value
	currentCounterValue, err := kv.ReadInt(ctx, "global-counter")
	if err != nil {
		return 0, err
	}

	// Update value
	newCounterValue := currentCounterValue + delta
	err = kv.CompareAndSwap(ctx, "global-counter", currentCounterValue, newCounterValue, true)
	if err != nil {
		return updateValue(ctx, kv, delta)
	}

	return newCounterValue, nil
}
