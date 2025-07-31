package correlation

type (
	NodeType int
	Status   int
)

const (
	FiberNode NodeType = iota
	SplitterNode
	SegmentNode
	CEONode
	CTONode
	CONode
	SensorNode
	ONUNode
)

const (
	Active Status = iota
	Alarmed
	ProbablyAlarmed
	Undefined
	Inconsistent
)

var nodeName = map[NodeType]string{
	FiberNode:    "Fiber",
	SplitterNode: "Splitter",
	SegmentNode:  "Segment",
	CEONode:      "CEO",
	CTONode:      "CTO",
	CONode:       "CO",
	SensorNode:   "Sensor",
	ONUNode:      "ONU",
}

var statusName = map[Status]string{
	Active:          "Active",
	Alarmed:         "Alarmed",
	ProbablyAlarmed: "Probably_Alarmed",
	Undefined:       "Undefined",
	Inconsistent:    "Inconsistent",
}

func (nt NodeType) String() string {
	return nodeName[nt]
}

func (s Status) String() string {
	return statusName[s]
}

type Node struct {
	ID       string
	Name     string
	Type     NodeType
	Status   Status
	Children []*Node
	Parents  []*Node
}

func NewNode(id string, name string, nodeType NodeType) *Node {
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
