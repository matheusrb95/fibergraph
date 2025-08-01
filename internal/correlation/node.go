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
	DIONode
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
	FiberNode:    "FIBER",
	SplitterNode: "SPLITTER",
	SegmentNode:  "SEGMENT",
	CEONode:      "CEO",
	CTONode:      "CTO",
	CONode:       "CO",
	DIONode:      "DIO",
	SensorNode:   "SENSOR",
	ONUNode:      "ONU",
}

var statusName = map[Status]string{
	Active:          "ACTIVE",
	Alarmed:         "ALARMED",
	ProbablyAlarmed: "PROBABLY_ALARMED",
	Undefined:       "UNDEFINED",
	Inconsistent:    "INCONSISTENT",
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

func (n *Node) ActiveSensor() bool {
	return n.Type == SensorNode && n.Status == Active
}

func (n *Node) AlarmedSensor() bool {
	return n.Type == SensorNode && n.Status == Alarmed
}
