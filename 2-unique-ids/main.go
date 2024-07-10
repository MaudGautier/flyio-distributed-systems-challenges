package main

import (
	"encoding/json"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"strconv"
)

func main() {
	n := maelstrom.NewNode()

	var lastId uint64 = 0

	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		var newId uint64 = lastId + 1
		body["id"] = n.ID() + "_" + strconv.FormatUint(newId, 10)

		lastId = lastId + 1

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
