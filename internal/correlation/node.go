package correlation

type (
	NodeType int
	Status   int
)

const (
	CTONode NodeType = iota
	CEONode
	CONode
	FiberNode
	SplitterNode
	SegmentNode
	SensorNode
	ONUNode
)

const (
	Active Status = iota
	Alarmed
	ProbablyAlarmed
	Undefined
)

var nodeName = map[NodeType]string{
	CTONode:      "CTO",
	CEONode:      "CEO",
	CONode:       "CO",
	FiberNode:    "Fiber",
	SplitterNode: "Splitter",
	SegmentNode:  "Segment",
	SensorNode:   "Sensor",
	ONUNode:      "ONU",
}

var statusName = map[Status]string{
	Active:          "Active",
	Alarmed:         "Alarmed",
	ProbablyAlarmed: "Probably_Alarmed",
	Undefined:       "Undefined",
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
	Parents  []*Node
}

func NewNode(id int, name string, nodeType NodeType) *Node {
	return &Node{
		ID:     id,
		Name:   name,
		Type:   nodeType,
		Status: Undefined,
	}
}

func (n *Node) SetParents(nodes ...*Node) {
	n.Parents = append(n.Parents, nodes...)
	for _, node := range nodes {
		node.addChildren(n)
	}
}

func (n *Node) SetChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
	for _, node := range nodes {
		node.addParents(n)
	}
}

func (n *Node) addParents(nodes ...*Node) {
	n.Parents = append(n.Parents, nodes...)
}

func (n *Node) addChildren(nodes ...*Node) {
	n.Children = append(n.Children, nodes...)
}
