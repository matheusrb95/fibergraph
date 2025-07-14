package data

type NodeType int

const (
	OLTNodeType NodeType = iota
	ONUNodeType
	CableNodeType
	BoxNodeType
	FiberNodeType
	SplitterNodeType
	SegmentNodeType
)

var nodeName = map[NodeType]string{
	OLTNodeType:      "OLT",
	ONUNodeType:      "ONU",
	CableNodeType:    "Cable",
	BoxNodeType:      "Box",
	FiberNodeType:    "Fiber",
	SplitterNodeType: "Splitter",
	SegmentNodeType:  "Segment",
}

func (nt NodeType) String() string {
	return nodeName[nt]
}

var id int

func ID() int {
	id++
	return id
}

type Node struct {
	ID        int
	Name      string
	Type      NodeType
	Active    bool
	RootCause bool
	Children  []*Node
	Parent    *Node
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
	n.Parent = node
	node.addChildren(n)
}

func (n *Node) SetChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
	for _, node := range nodes {
		node.addParent(n)
	}
}

func (n *Node) addChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
}

func (n *Node) addParent(node *Node) {
	n.Parent = node
}

func (n *Node) Depth() int {
	var result int
	for n.Parent != nil {
		n = n.Parent
		result++
	}

	return result
}
