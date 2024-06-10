package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"os"
)

func main() {
	n := maelstrom.NewNode()

	// Broadcast - input message body
	//{
	//  "type": "broadcast",
	//  "message": 1000
	//}
	//
	// Broadcast - response
	//{
	//  "type": "broadcast_ok",
	//}

	var messages []interface{}
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Add message to list of messages
		messages = append(messages, body["message"])

		// Create return body
		returnBody := make(map[string]interface{})
		returnBody["type"] = "broadcast_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, returnBody)
	})

	// Read - input message body
	//{
	//  "type": "read"
	//}
	//
	// Read - response
	//{
	//  "type": "read_ok",
	//  "messages": [1, 8, 72, 25]
	//}

	n.Handle("read", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Create return body
		returnBody := make(map[string]interface{})
		returnBody["type"] = "read_ok"
		returnBody["messages"] = messages

		// Echo the original message back with the updated message type.
		return n.Reply(msg, returnBody)
	})

	// Topology - input message body
	//{
	//  "type": "topology",
	//  "topology": {
	//    "n1": ["n2", "n3"],
	//    "n2": ["n1"],
	//    "n3": ["n1"]
	//  }
	//}
	//
	// Topology - response
	//{
	//  "type": "topology_ok"
	//}

	n.Handle("topology", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			fmt.Println("ERROR is ", err)
			return err
		}

		fmt.Fprintf(os.Stderr, "Get to return body\n")

		// Create return body
		returnBody := make(map[string]string)
		returnBody["type"] = "topology_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, returnBody)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
