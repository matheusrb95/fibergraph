package data

type (
	NodeType int
	Status   int
)

const (
	BoxNode NodeType = iota
	FiberNode
	SplitterNode
	SegmentNode
	SensorNode
)

const (
	Active Status = iota
	Alarmed
	Unknown
)

var nodeName = map[NodeType]string{
	BoxNode:      "Box",
	FiberNode:    "Fiber",
	SplitterNode: "Splitter",
	SegmentNode:  "Segment",
	SensorNode:   "Sensor",
}

var statusName = map[Status]string{
	Active:  "Active",
	Alarmed: "Alarmed",
	Unknown: "Unknown",
}

func (nt NodeType) String() string {
	return nodeName[nt]
}

func (s Status) String() string {
	return statusName[s]
}

type Node struct {
	ID       int
	Name     string
	Type     NodeType
	Status   Status
	Children []*Node
	Parent   *Node
}

func NewNode(id int, name string, nodeType NodeType) *Node {
	return &Node{
		ID:     id,
		Name:   name,
		Type:   nodeType,
		Status: Unknown,
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

func (n *Node) addParent(node *Node) {
	n.Parent = node
}

func (n *Node) addChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
}
