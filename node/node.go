package node

type (
	Node struct {
		state state
	}
	state interface{}
)

func New(s state) *Node {
	return &Node{s}
}
