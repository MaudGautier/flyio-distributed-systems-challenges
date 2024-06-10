package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

func main() {
	n := maelstrom.NewNode()

	var messages []interface{}

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

		// Broadcast to other nodes in the topology
		neighbors := n.NodeIDs()

		for _, neighbor := range neighbors {
			neighborBody := make(map[string]interface{})
			neighborBody["type"] = "rebroadcast"
			neighborBody["message"] = body["message"]
			n.Send(neighbor, neighborBody)
		}

		// Echo the original message back with the updated message type.
		return n.Reply(msg, returnBody)
	})

	n.Handle("rebroadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Add message to list of messages
		messages = append(messages, body["message"])

		// Echo the original message back with the updated message type.
		return nil
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
			return err
		}

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
