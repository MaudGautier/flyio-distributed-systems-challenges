package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
	"time"
)

// Message specification
//{
//  "src": "c1",
//  "dest": "n1",
//  "body": {
//    "type": "echo",
//    "msg_id": 1,
//    "echo": "Please echo 35"
//  }
//}

var mutex sync.Mutex

func main() {

	n := maelstrom.NewNode()

	var messages []interface{}

	go periodicBroadcast(n, &messages)

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

		// Create return body
		returnBody := make(map[string]interface{})
		returnBody["type"] = "broadcast_ok"

		// If message already seen, do nothing (only reply ok)
		if seen := isMessageInList(messages, body["message"]); seen {
			return n.Reply(msg, returnBody)
		}

		// Add message to list of messages
		mutex.Lock()
		messages = append(messages, body["message"])
		mutex.Unlock()

		// Rebroadcast to all other nodes
		neighbors := getFlatTreeNeighbors(n)

		for _, neighbor := range neighbors {
			body["type"] = "rebroadcast"
			n.Send(neighbor, body)
		}

		// Echo the original message back with the updated message type.
		return n.Reply(msg, returnBody)
	})

	n.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		return nil
	})

	n.Handle("rebroadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// If message already seen, do nothing (only reply ok)
		if seen := isMessageInList(messages, body["message"]); seen {
			return nil
		}

		// Add message to list of messages
		mutex.Lock()
		messages = append(messages, body["message"])
		mutex.Unlock()

		return nil
	})

	n.Handle("periodic_broadcast", func(msg maelstrom.Message) error {
		// Unmarshal the message body as a loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		for _, message := range body["message"].([]interface{}) {

			// If message already seen, do nothing
			if seen := isMessageInList(messages, message); seen {
				continue
			}

			// Add message to list of messages
			mutex.Lock()
			messages = append(messages, message)
			mutex.Unlock()

		}

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

func getFlatTreeNeighbors(n *maelstrom.Node) []string {
	var allNeighbors []string
	for _, id := range n.NodeIDs() {
		if id != n.ID() {
			allNeighbors = append(allNeighbors, id)
		}
	}
	return allNeighbors
}

func isMessageInList(messages []interface{}, searchedMessage interface{}) bool {
	for _, message := range messages {
		if message == searchedMessage {
			return true
		}
	}
	return false
}

func periodicBroadcast(node *maelstrom.Node, messages *[]interface{}) {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		// Broadcast to other nodes in the topology
		neighbors := getFlatTreeNeighbors(node)

		for _, neighbor := range neighbors {
			if neighbor == node.ID() {
				continue
			}

			// Create a new message body from scratch with all messages
			body := map[string]interface{}{
				"type":    "periodic_broadcast",
				"message": messages,
			}
			node.Send(neighbor, body)

		}

	}

}
