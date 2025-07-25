package correlation

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/matheusrb95/fibergraph/internal/data"
)

type Correlation struct {
	Connections    []*data.Connection
	Sensors        []*data.Sensor
	ONUs           []*data.ONU
	ActiveSensors  []string
	AlarmedSensors []string
	ActiveONUs     []string
	AlarmedONUs    []string
	Segments       []*data.Segment
	Components     []*data.Component

	nodes         map[int]*Node
	nodesToUpdate []*Node
}

func New(
	connections []*data.Connection,
	sensors []*data.Sensor,
	onus []*data.ONU,
	activeSensors []string,
	alarmedSensors []string,
	activeONUs []string,
	alarmedONUs []string,
	segments []*data.Segment,
	components []*data.Component,
) *Correlation {
	return &Correlation{
		Connections:    connections,
		Sensors:        sensors,
		ONUs:           onus,
		ActiveSensors:  activeSensors,
		AlarmedSensors: alarmedSensors,
		ActiveONUs:     activeONUs,
		AlarmedONUs:    alarmedONUs,
		Segments:       segments,
		Components:     components,
		nodes:          make(map[int]*Node),
		nodesToUpdate:  make([]*Node, 0),
	}
}

func (c *Correlation) Result() []*Node {
	return c.nodesToUpdate
}

func (c *Correlation) Run() error {
	rootNodes := c.BuildNetworkWithConnection()

	if len(rootNodes) == 0 {
		return errors.New("no nodes")
	}

	for _, rootNode := range rootNodes {
		propagateSensorStatus(rootNode)
		propagateONUStatus(rootNode)
		err := drawGraphs(rootNode)
		if err != nil {
			return err
		}
	}

	segmentNodes := c.BuildSegmentNodes()
	rootNodes = c.BuildComponentNodes(segmentNodes)

	if os.Getenv("DRAW_CORRELATION") != "true" {
		return nil
	}

	if len(rootNodes) == 0 {
		return errors.New("no nodes")
	}

	for _, rootNode := range rootNodes {
		err := drawGraphs(rootNode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Correlation) BuildNetworkWithConnection() []*Node {
	result := make([]*Node, 0)

	for _, connection := range c.Connections {
		c.upsertConnectionMap(connection)

		switch connection.Type {
		case "CO":
			result = append(result, c.nodes[connection.ID])
		case "Splitter":
		case "ONU":
			c.nodesToUpdate = append(c.nodesToUpdate, c.nodes[connection.ID])
		}
	}

	for _, sensor := range c.Sensors {
		fiberNode, ok := c.nodes[sensor.FiberID]
		if !ok {
			continue
		}

		name := fmt.Sprintf("%d - SPD", sensor.ID)
		node := NewNode(sensor.ID+1_000_000, name, SensorNode)
		var status Status

		if slices.Contains(c.AlarmedSensors, sensor.DevEUI) {
			status = Alarmed
		} else if slices.Contains(c.ActiveSensors, sensor.DevEUI) {
			status = Active
		} else {
			status = Undefined
		}
		// switch sensor.Status {
		// case "ACTIVE":
		// status = Active
		// case "ALARMED":
		// status = Alarmed
		// default:
		// status = Undefined
		// }
		node.Status = status
		node.SetParents(fiberNode)

		c.nodesToUpdate = append(c.nodesToUpdate, node)
	}

	for _, onu := range c.ONUs {
		onuNode, ok := c.nodes[onu.ID]
		if !ok {
			continue
		}

		var status Status
		if slices.Contains(c.AlarmedONUs, onu.Serial) {
			status = Alarmed
		} else if slices.Contains(c.ActiveONUs, onu.Serial) {
			status = Active
		} else {
			status = Undefined
		}

		onuNode.Type = ONUNode
		onuNode.Status = status
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
			parentID, err := strconv.Atoi(parentID)
			if err != nil {
				continue
			}

			parentNode, ok := c.nodes[parentID]
			if !ok {
				continue
			}

			node := c.nodes[connection.ID]
			node.SetParents(parentNode)
		}
	}

	return result
}

func (c *Correlation) BuildSegmentNodes() map[int]*Node {
	result := make(map[int]*Node)

	for _, segment := range c.Segments {
		if segment.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%d - segment", segment.ID)
		segmentNode := NewNode(segment.ID, name, SegmentNode)

		var hasActive, hasAlarmed, hasProbablyAlarmed, hasUndefined bool

		fiberIDs := strings.SplitSeq(*segment.FiberIDs, ",")
		for fiberID := range fiberIDs {
			fiberID, err := strconv.Atoi(fiberID)
			if err != nil {
				continue
			}

			node, ok := c.nodes[fiberID]
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

		result[segment.ID] = segmentNode
		c.nodesToUpdate = append(c.nodesToUpdate, segmentNode)
	}

	return result
}

func (c *Correlation) BuildComponentNodes(segmentNodes map[int]*Node) []*Node {
	result := make([]*Node, 0)

	for _, component := range c.Components {
		name := fmt.Sprintf("%d - %s", component.ID, component.Name)
		var nodeType NodeType
		switch component.Type {
		case "CO":
			nodeType = CONode
		case "CEO":
			nodeType = CEONode
		case "CTO":
			nodeType = CTONode
		case "ONU":
			nodeType = ONUNode
		}
		componentNode := NewNode(component.ID, name, nodeType)

		childrenIDs := []string{}
		if component.ChildrenIDs != nil {
			childrenIDs = strings.Split(*component.ChildrenIDs, ",")
		}
		for _, childID := range childrenIDs {
			childID, err := strconv.Atoi(childID)
			if err != nil {
				continue
			}

			if _, ok := segmentNodes[childID]; !ok {
				slog.Warn("segment node does not exist")
				continue
			}

			componentNode.SetChildren(segmentNodes[childID])
		}

		var hasActive, hasAlarmed, hasProbablyAlarmed, hasUndefined bool

		parentIDs := []string{}
		if component.ParentIDs != nil {
			parentIDs = strings.Split(*component.ParentIDs, ",")
		}
		for _, parentID := range parentIDs {
			parentID, err := strconv.Atoi(parentID)
			if err != nil {
				continue
			}

			if _, ok := segmentNodes[parentID]; !ok {
				slog.Warn("segment node does not exist")
				continue
			}

			componentNode.SetParents(segmentNodes[parentID])

			switch segmentNodes[parentID].Status {
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

		if component.ParentIDs == nil {
			result = append(result, componentNode)
			continue
		}
		c.nodesToUpdate = append(c.nodesToUpdate, componentNode)
	}

	return result
}

func (c *Correlation) upsertConnectionMap(connection *data.Connection) {
	if _, ok := c.nodes[connection.ID]; ok {
		return
	}

	name := fmt.Sprintf("%d - %s", connection.ID, connection.Name)
	var nodeType NodeType
	switch connection.Type {
	case "CO":
		nodeType = CONode
	case "CTO":
		nodeType = CTONode
	case "ONU":
		nodeType = ONUNode
	case "Fiber":
		nodeType = FiberNode
	case "Splitter":
		nodeType = SplitterNode
	case "Sensor":
		nodeType = SensorNode
	}
	c.nodes[connection.ID] = NewNode(connection.ID, name, nodeType)
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
		if child.Type == SensorNode && child.Status == Active {
			break
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
