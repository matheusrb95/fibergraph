package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/matheusrb95/fibergraph/internal/core"
	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/response"
)

func HandleDraw(logger *slog.Logger, models *data.Models) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")

		connections, err := models.Connection.GetAll(tenantID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		sensors, err := models.Sensor.GetAll(tenantID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		nodes := make(map[int]*data.Node)

		rootNodes := buildNetworkWithConnection(nodes, connections, sensors)

		if len(rootNodes) == 0 {
			serverErrorResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = core.Run(rootNode, true)
			if err != nil {
				serverErrorResponse(w, r, logger, err)
				return
			}
		}

		segments, err := models.Segment.GetAll(tenantID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		segmentNodes := buildSegmentNodes(nodes, segments)

		components, err := models.Component.GetAll(tenantID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		rootNodes = buildComponentNodes(components, segmentNodes)

		if len(rootNodes) == 0 {
			serverErrorResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = core.Run(rootNode, true)
			if err != nil {
				serverErrorResponse(w, r, logger, err)
				return
			}
		}

		err = response.JSON(w, http.StatusCreated, response.Envelope{"message": "draw done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}

func buildNetworkWithConnection(nodes map[int]*data.Node, connections []*data.Connection, sensors []*data.Sensor) []*data.Node {
	result := make([]*data.Node, 0)

	for _, connection := range connections {
		upsertConnectionMap(nodes, connection)

		if connection.ParentIDs == nil {
			result = append(result, nodes[connection.ID])
		}
	}

	for _, sensor := range sensors {
		fiberNode, ok := nodes[sensor.FiberID]
		if !ok {
			continue
		}

		name := fmt.Sprintf("%d - SPD", sensor.ID)
		node := data.NewNode(sensor.ID+5000, name, data.SensorNode)
		var status data.Status
		switch sensor.Status {
		case "ACTIVE":
			status = data.Active
		case "ALARMED":
			status = data.Inactive
		default:
			status = data.Unknown
		}
		node.Status = status
		node.SetParent(fiberNode)
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
		node.SetParent(parentNode)
	}

	return result
}

func upsertConnectionMap(nodes map[int]*data.Node, connection *data.Connection) {
	if _, ok := nodes[connection.ID]; ok {
		return
	}

	name := fmt.Sprintf("%d - %s", connection.ID, connection.Name)
	nodes[connection.ID] = data.NewNode(connection.ID, name, data.BoxNode)
}

func buildSegmentNodes(nodes map[int]*data.Node, segments []*data.Segment) map[int]*data.Node {
	result := make(map[int]*data.Node)

	for _, segment := range segments {
		if segment.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%d - segment", segment.ID)
		segmentNode := data.NewNode(segment.ID, name, data.SegmentNode)

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
			case data.Active:
				hasActive = true
			case data.Inactive:
				hasInactive = true
			case data.Unknown:
				hasUnknown = true
			}
		}

		switch {
		case hasActive:
			segmentNode.Status = data.Active
		case hasInactive:
			segmentNode.Status = data.Inactive
		case hasUnknown:
			segmentNode.Status = data.Unknown
		}

		result[segment.ID] = segmentNode
	}

	return result
}

func buildComponentNodes(components []*data.Component, segmentNodes map[int]*data.Node) []*data.Node {
	result := make([]*data.Node, 0)

	for _, component := range components {
		name := fmt.Sprintf("%d - component", component.ID)
		componentNode := data.NewNode(component.ID, name, data.BoxNode)

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

			componentNode.SetParent(segmentNodes[parentID])

			if componentNode.Children == nil {
				switch segmentNodes[parentID].Status {
				case data.Active:
					hasActive = true
				case data.Inactive:
					hasInactive = true
				case data.Unknown:
					hasUnknown = true
				}
			}
		}

		switch {
		case hasActive:
			componentNode.Status = data.Active
		case hasInactive:
			componentNode.Status = data.Inactive
		case hasUnknown:
			componentNode.Status = data.Unknown
		}

		if component.ParentIDs == nil {
			result = append(result, componentNode)
		}
	}

	return result
}
