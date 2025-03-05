package node

type Node struct{}

func New() *Node {
	return new(Node)
}
