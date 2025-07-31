package correlation

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/matheusrb95/fibergraph/internal/data"
)

type Correlation struct {
	Connections     []*data.Connection
	Sensors         []*data.Sensor
	ONUs            []*data.ONU
	ActiveSensors   []string
	AlarmedSensors  []string
	InactiveSensors []string
	ActiveONUs      []string
	AlarmedONUs     []string
	Segments        []*data.Segment
	Components      []*data.Component

	connectionNodes map[string]*Node
	topologicNodes  []*Node
}

func New(
	connections []*data.Connection,
	sensors []*data.Sensor,
	onus []*data.ONU,
	activeSensors []string,
	alarmedSensors []string,
	inactiveSensors []string,
	activeONUs []string,
	alarmedONUs []string,
	segments []*data.Segment,
	components []*data.Component,
) *Correlation {
	return &Correlation{
		Connections:     connections,
		Sensors:         sensors,
		ONUs:            onus,
		ActiveSensors:   activeSensors,
		AlarmedSensors:  alarmedSensors,
		InactiveSensors: inactiveSensors,
		ActiveONUs:      activeONUs,
		AlarmedONUs:     alarmedONUs,
		Segments:        segments,
		Components:      components,
		connectionNodes: make(map[string]*Node),
		topologicNodes:  make([]*Node, 0),
	}
}

func (c *Correlation) Result() []*Node {
	return c.topologicNodes
}

func (c *Correlation) Run() error {
	rootNodes := c.buildNetworkWithConnection()

	if len(rootNodes) == 0 {
		return errors.New("no nodes")
	}

	for _, rootNode := range rootNodes {
		inconsistentSensors := make([]*Node, 0)
		checkInconsistentSensor(rootNode, &inconsistentSensors)
		if len(inconsistentSensors) != 0 {
			c.determineInconsistentSensor(inconsistentSensors[1], inconsistentSensors[0])
		}

		propagateSensorStatus(rootNode)
		propagateONUStatus(rootNode)

		if os.Getenv("DRAW_CORRELATION") == "true" {
			err := drawGraphs(rootNode)
			if err != nil {
				return err
			}
		}
	}

	c.determineSegmentsStatus()
	c.determineComponentsStatus()

	return nil
}

func (c *Correlation) determineInconsistentSensor(alarmedNode, activeNode *Node) {
	alarmedInList := slices.Contains(c.AlarmedSensors, alarmedNode.ID)
	activeInList := slices.Contains(c.ActiveSensors, activeNode.ID)

	switch {
	case activeInList && !alarmedInList:
		alarmedNode.Status = Inconsistent
	case alarmedInList && !activeInList:
		activeNode.Status = Inconsistent
	default:
		alarmedNode.Status = Inconsistent
	}
}

func (c *Correlation) buildNetworkWithConnection() []*Node {
	result := make([]*Node, 0)

	for _, connection := range c.Connections {
		c.updateConnectionMap(connection)

		switch connection.Type {
		case "CO":
			result = append(result, c.connectionNodes[connection.ID])
		case "Splitter":
			c.topologicNodes = append(c.topologicNodes, c.connectionNodes[connection.ID])
		}
	}

	for _, sensor := range c.Sensors {
		fiberNode, ok := c.connectionNodes[sensor.FiberID]
		if !ok {
			continue
		}

		name := fmt.Sprintf("%s - SPD", sensor.DevEUI)
		node := NewNode(sensor.DevEUI, name, SensorNode)

		var status Status
		switch sensor.Status {
		case "ACTIVE":
			status = Active
		case "ALARMED":
			status = Alarmed
		default:
			status = Undefined
		}

		if slices.Contains(c.AlarmedSensors, sensor.DevEUI) {
			status = Alarmed
		} else if slices.Contains(c.ActiveSensors, sensor.DevEUI) {
			status = Active
		} else if slices.Contains(c.InactiveSensors, sensor.DevEUI) {
			status = Undefined
		} else {
			status = Undefined
		}

		node.Status = status
		node.SetParents(fiberNode)

		c.topologicNodes = append(c.topologicNodes, node)
	}

	for _, onu := range c.ONUs {
		fiberNode, ok := c.connectionNodes[onu.FiberID]
		if !ok {
			continue
		}

		name := fmt.Sprintf("%s - ONU", onu.SerialNumber)
		node := NewNode(onu.SerialNumber, name, ONUNode)

		var status Status
		if slices.Contains(c.AlarmedONUs, onu.SerialNumber) {
			status = Alarmed
		} else if slices.Contains(c.ActiveONUs, onu.SerialNumber) {
			status = Active
		} else {
			status = Undefined
		}

		node.Status = status
		node.SetParents(fiberNode)

		c.topologicNodes = append(c.topologicNodes, node)
	}

	for _, connection := range c.Connections {
		if connection.ParentIDs == nil {
			continue
		}

		parentIDs := []string{}
		if connection.ParentIDs != nil {
			parentIDs = strings.Split(*connection.ParentIDs, ",")
		}
		for _, parentID := range parentIDs {
			parentNode, ok := c.connectionNodes[parentID]
			if !ok {
				continue
			}

			node := c.connectionNodes[connection.ID]
			node.SetParents(parentNode)
		}
	}

	return result
}

func (c *Correlation) determineSegmentsStatus() {
	for _, segment := range c.Segments {
		if segment.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%s - segment", segment.ID)
		segmentNode := NewNode(segment.ID, name, SegmentNode)

		var hasActive, hasAlarmed, hasProbablyAlarmed, hasUndefined bool

		fiberIDs := strings.SplitSeq(*segment.FiberIDs, ",")
		for fiberID := range fiberIDs {
			node, ok := c.connectionNodes[fiberID]
			if !ok {
				continue
			}

			switch node.Status {
			case Active:
				hasActive = true
			case Alarmed:
				hasAlarmed = true
			case ProbablyAlarmed:
				hasProbablyAlarmed = true
			case Undefined:
				hasUndefined = true
			}
		}

		switch {
		case hasActive:
			segmentNode.Status = Active
		case hasAlarmed:
			segmentNode.Status = Alarmed
		case hasProbablyAlarmed:
			segmentNode.Status = ProbablyAlarmed
		case hasUndefined:
			segmentNode.Status = Undefined
		}

		c.topologicNodes = append(c.topologicNodes, segmentNode)
	}
}

func (c *Correlation) determineComponentsStatus() {
	for _, component := range c.Components {
		if component.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%s - component", component.ID)
		var nodeType NodeType
		switch component.Type {
		case "CEO":
			nodeType = CEONode
		case "CTO":
			nodeType = CTONode
		case "CO":
			nodeType = CONode
		}
		componentNode := NewNode(component.ID, name, nodeType)

		var hasActive, hasAlarmed, hasProbablyAlarmed, hasUndefined bool

		fiberIDs := strings.SplitSeq(*component.FiberIDs, ",")
		for fiberID := range fiberIDs {
			node, ok := c.connectionNodes[fiberID]
			if !ok {
				continue
			}

			switch node.Status {
			case Active:
				hasActive = true
			case Alarmed:
				hasAlarmed = true
			case ProbablyAlarmed:
				hasProbablyAlarmed = true
			case Undefined:
				hasUndefined = true
			}
		}

		switch {
		case hasActive:
			componentNode.Status = Active
		case hasAlarmed:
			componentNode.Status = Alarmed
		case hasProbablyAlarmed:
			componentNode.Status = ProbablyAlarmed
		case hasUndefined:
			componentNode.Status = Undefined
		}
		c.topologicNodes = append(c.topologicNodes, componentNode)
	}
}

func (c *Correlation) updateConnectionMap(connection *data.Connection) {
	if _, ok := c.connectionNodes[connection.ID]; ok {
		return
	}

	name := fmt.Sprintf("%s - %s", connection.ID, connection.Name)
	var nodeType NodeType
	switch connection.Type {
	case "CEO":
		nodeType = CEONode
	case "CTO":
		nodeType = CTONode
	case "CO":
		nodeType = CONode
	case "ONU":
		nodeType = ONUNode
	case "Fiber":
		nodeType = FiberNode
	case "Splitter":
		nodeType = SplitterNode
	case "Sensor":
		nodeType = SensorNode
	}
	c.connectionNodes[connection.ID] = NewNode(connection.ID, name, nodeType)
}

func checkInconsistentSensor(node *Node, result *[]*Node) {
	if node.Children == nil {
		return
	}

	for _, child := range node.Children {
		if child.Type == SensorNode && child.Status == Alarmed {
			if hasSensorActiveBelow(node, result) {
				*result = append(*result, child)
			}
		}

		checkInconsistentSensor(child, result)
	}
}

func hasSensorActiveBelow(node *Node, result *[]*Node) bool {
	for _, child := range node.Children {
		if child.Type == SensorNode {
			switch child.Status {
			case Active:
				*result = append(*result, child)
				return true
			default:
				continue
			}
		}

		if hasSensorActiveBelow(child, result) {
			return true
		}
	}

	return false
}

func propagateSensorStatus(node *Node) {
	if node.Children == nil {
		return
	}

	for _, child := range node.Children {
		propagateSensorStatus(child)

		if child.Type != SensorNode {
			continue
		}

		switch child.Status {
		case Active:
			activeAllAbove(node)
		case Alarmed:
			alarmAllBelow(node)
			probablyAlarmedAllAbove(node)
		}
	}
}

func propagateONUStatus(node *Node) {
	if node.Children == nil {
		return
	}

	for _, child := range node.Children {
		propagateONUStatus(child)

		if child.Type != ONUNode {
			continue
		}

		switch child.Status {
		case Active:
			activeAllAbove(node)
		case Alarmed:
			alarmAllBelow(node)
			probablyAlarmedAllAbove(node)
		}
	}
}

func activeAllAbove(node *Node) {
	if node.Parents == nil {
		return
	}

	node.Status = Active

	for _, parent := range node.Parents {
		activeAllAbove(parent)
		parent.Status = Active
	}
}

func alarmAllBelow(node *Node) {
	if node.Children == nil {
		return
	}

	node.Status = Alarmed

	for _, child := range node.Children {
		alarmAllBelow(child)
		if child.Type == SensorNode {
			continue
		}

		child.Status = Alarmed
	}
}

func probablyAlarmedAllAbove(node *Node) {
	if node.Parents == nil {
		return
	}

	for _, parent := range node.Parents {
		probablyAlarmedAllAbove(parent)
		if parent.Status != Undefined {
			continue
		}

		parent.Status = ProbablyAlarmed
	}
}
