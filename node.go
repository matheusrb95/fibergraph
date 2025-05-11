package main

type NodeType int

const (
	OLTNodeType NodeType = iota
	ONUNodeType
	CableNodeType
	BoxNodeType
	SplitterNodeType
)

var nodeName = map[NodeType]string{
	OLTNodeType:      "OLT",
	ONUNodeType:      "ONU",
	CableNodeType:    "Cable",
	BoxNodeType:      "Box",
	SplitterNodeType: "Splitter",
}

func (nt NodeType) String() string {
	return nodeName[nt]
}

type Node struct {
	ID        int
	Name      string
	Type      NodeType
	Active    bool
	RootCause bool
	children  []*Node
	parent    *Node
}

func NewNode(id int, name string, nodeType NodeType) *Node {
	return &Node{
		ID:     id,
		Name:   name,
		Type:   nodeType,
		Active: true,
	}
}

func (n *Node) SetParent(node *Node) {
	n.parent = node
}

func (n *Node) SetChildren(nodes ...*Node) {
	n.children = append(n.children, nodes...)
}

func (n *Node) Depth() int {
	var result int
	for n.parent != nil {
		n = n.parent
		result++
	}

	return result
}
