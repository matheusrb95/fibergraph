package correlation

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/matheusrb95/fibergraph/internal/data"
)

type Correlation struct {
	nodes map[int]*Node
}

func New() *Correlation {
	return &Correlation{}
}

func BuildNetworkWithConnection(
	nodes map[int]*Node,
	connections []*data.Connection,
	sensors []*data.Sensor,
	activeSensors []string,
	alarmedSensors []string,
) []*Node {
	result := make([]*Node, 0)

	for _, connection := range connections {
		upsertConnectionMap(nodes, connection)

		if connection.CentralOffice {
			result = append(result, nodes[connection.ID])
		}
	}

	for _, sensor := range sensors {
		fiberNode, ok := nodes[sensor.FiberID]
		if !ok {
			continue
		}

		name := fmt.Sprintf("%d - SPD", sensor.ID)
		node := NewNode(sensor.ID+1_000_000, name, SensorNode)
		var status Status

		if slices.Contains(alarmedSensors, sensor.DevEUI) {
			status = Alarmed
		} else if slices.Contains(activeSensors, sensor.DevEUI) {
			status = Active
		} else {
			status = Unknown
		}
		// switch sensor.Status {
		// case "ACTIVE":
		// status = Active
		// case "ALARMED":
		// status = Alarmed
		// default:
		// status = Unknown
		// }
		node.Status = status
		node.SetParents(fiberNode)
	}

	for _, connection := range connections {
		if connection.ParentIDs == nil {
			continue
		}

		parentID, err := strconv.Atoi(*connection.ParentIDs)
		if err != nil {
			continue
		}

		parentNode, ok := nodes[parentID]
		if !ok {
			continue
		}
		node := nodes[connection.ID]
		node.SetParents(parentNode)
	}

	return result
}

func upsertConnectionMap(nodes map[int]*Node, connection *data.Connection) {
	if _, ok := nodes[connection.ID]; ok {
		return
	}

	name := fmt.Sprintf("%d - %s", connection.ID, connection.Name)
	nodes[connection.ID] = NewNode(connection.ID, name, BoxNode)
}

func BuildSegmentNodes(nodes map[int]*Node, segments []*data.Segment) map[int]*Node {
	result := make(map[int]*Node)

	for _, segment := range segments {
		if segment.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%d - segment", segment.ID)
		segmentNode := NewNode(segment.ID, name, SegmentNode)

		var hasActive, hasInactive, hasUnknown bool

		fiberIDs := strings.SplitSeq(*segment.FiberIDs, ",")
		for fiberID := range fiberIDs {
			fiberID, err := strconv.Atoi(fiberID)
			if err != nil {
				continue
			}

			node, ok := nodes[fiberID]
			if !ok {
				continue
			}

			switch node.Status {
			case Active:
				hasActive = true
			case Alarmed:
				hasInactive = true
			case Unknown:
				hasUnknown = true
			}
		}

		switch {
		case hasActive:
			segmentNode.Status = Active
		case hasInactive:
			segmentNode.Status = Alarmed
		case hasUnknown:
			segmentNode.Status = Unknown
		}

		result[segment.ID] = segmentNode
	}

	return result
}

func BuildComponentNodes(components []*data.Component, segmentNodes map[int]*Node) []*Node {
	result := make([]*Node, 0)

	for _, component := range components {
		name := fmt.Sprintf("%d - %s", component.ID, component.Name)
		componentNode := NewNode(component.ID, name, BoxNode)

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

		var hasActive, hasInactive, hasUnknown bool

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
				hasInactive = true
			case Unknown:
				hasUnknown = true
			}
		}

		switch {
		case hasActive:
			componentNode.Status = Active
		case hasInactive:
			componentNode.Status = Alarmed
		case hasUnknown:
			componentNode.Status = Unknown
		}

		if component.ParentIDs == nil {
			result = append(result, componentNode)
		}
	}

	return result
}
