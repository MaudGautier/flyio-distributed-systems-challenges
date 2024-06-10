package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
)

//{
//  "src": "c1",
//  "dest": "n1",
//  "body": {
//    "type": "echo",
//    "msg_id": 1,
//    "echo": "Please echo 35"
//  }
//}

//{
//  "src": "n1",
//  "dest": "c1",
//  "body": {
//    "type": "echo_ok",
//    "msg_id": 1,
//    "in_reply_to": 1,
//    "echo": "Please echo 35"
//  }
//}

func main() {
	n := maelstrom.NewNode()

	n.Handle("echo", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "echo_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	//fmt.Println("Starting server...")

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
