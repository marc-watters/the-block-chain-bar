package node

type (
	Node  struct{}
	state interface{}
)

func New() *Node {
	return new(Node)
}
