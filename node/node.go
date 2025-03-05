package node

import (
	"fmt"
	"net/http"
)

type (
	Node struct {
		state state
	}
	state interface{}
)

func New(s state) *Node {
	return &Node{s}
}

func (n *Node) Run() error {
	const port = 8080

	mx := http.NewServeMux()

	fmt.Printf("Listening on %s:%d", "127.0.0.1\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mx)
}
