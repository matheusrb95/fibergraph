package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/matheusrb95/fibergraph/internal/core"
	"github.com/matheusrb95/fibergraph/internal/correlation"
	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/request"
	"github.com/matheusrb95/fibergraph/internal/response"
)

type SensorStatus struct {
	Active  []string `json:"active_sensors"`
	Alarmed []string `json:"alarmed_sensors"`
}

func HandleDraw(logger *slog.Logger, models *data.Models) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.PathValue("tenant_id")
		if tenantID == "" {
			notFoundResponse(w, r, logger)
			return
		}

		projectID := r.PathValue("project_id")
		if projectID == "" {
			notFoundResponse(w, r, logger)
			return
		}

		var sensorStatus SensorStatus
		err := request.DecodeJSON(w, r, &sensorStatus)
		if err != nil {
			badRequestResponse(w, r, logger, err)
			return
		}

		connections, err := models.Connection.GetAll(tenantID, projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		sensors, err := models.Sensor.GetAll(tenantID, projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		nodes := make(map[int]*correlation.Node)

		rootNodes := buildNetworkWithConnection(nodes, connections, sensors, sensorStatus.Active, sensorStatus.Alarmed)

		if len(rootNodes) == 0 {
			badRequestResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = core.Run(rootNode, false, true, true)
			if err != nil {
				serverErrorResponse(w, r, logger, err)
				return
			}
		}

		segments, err := models.Segment.GetAll(tenantID, projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		segmentNodes := buildSegmentNodes(nodes, segments)

		components, err := models.Component.GetAll(tenantID, projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		logger.Info("network size",
			"connections_len", len(connections),
			"sensors_len", len(sensors),
			"segments_len", len(segments),
			"components_len", len(components),
		)

		rootNodes = buildComponentNodes(components, segmentNodes)

		if len(rootNodes) == 0 {
			badRequestResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = core.Run(rootNode, false, true, false)
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

func buildNetworkWithConnection(
	nodes map[int]*correlation.Node,
	connections []*data.Connection,
	sensors []*data.Sensor,
	activeSensors []string,
	alarmedSensors []string,
) []*correlation.Node {
	result := make([]*correlation.Node, 0)

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
		node := correlation.NewNode(sensor.ID+1_000_000, name, correlation.SensorNode)
		var status correlation.Status

		if slices.Contains(alarmedSensors, sensor.DevEUI) {
			status = correlation.Alarmed
		} else if slices.Contains(activeSensors, sensor.DevEUI) {
			status = correlation.Active
		} else {
			status = correlation.Unknown
		}
		// switch sensor.Status {
		// case "ACTIVE":
		// status = correlation.Active
		// case "ALARMED":
		// status = correlation.Alarmed
		// default:
		// status = correlation.Unknown
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

func upsertConnectionMap(nodes map[int]*correlation.Node, connection *data.Connection) {
	if _, ok := nodes[connection.ID]; ok {
		return
	}

	name := fmt.Sprintf("%d - %s", connection.ID, connection.Name)
	nodes[connection.ID] = correlation.NewNode(connection.ID, name, correlation.BoxNode)
}

func buildSegmentNodes(nodes map[int]*correlation.Node, segments []*data.Segment) map[int]*correlation.Node {
	result := make(map[int]*correlation.Node)

	for _, segment := range segments {
		if segment.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%d - segment", segment.ID)
		segmentNode := correlation.NewNode(segment.ID, name, correlation.SegmentNode)

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
			case correlation.Active:
				hasActive = true
			case correlation.Alarmed:
				hasInactive = true
			case correlation.Unknown:
				hasUnknown = true
			}
		}

		switch {
		case hasActive:
			segmentNode.Status = correlation.Active
		case hasInactive:
			segmentNode.Status = correlation.Alarmed
		case hasUnknown:
			segmentNode.Status = correlation.Unknown
		}

		result[segment.ID] = segmentNode
	}

	return result
}

func buildComponentNodes(components []*data.Component, segmentNodes map[int]*correlation.Node) []*correlation.Node {
	result := make([]*correlation.Node, 0)

	for _, component := range components {
		name := fmt.Sprintf("%d - %s", component.ID, component.Name)
		componentNode := correlation.NewNode(component.ID, name, correlation.BoxNode)

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
			case correlation.Active:
				hasActive = true
			case correlation.Alarmed:
				hasInactive = true
			case correlation.Unknown:
				hasUnknown = true
			}
		}

		switch {
		case hasActive:
			componentNode.Status = correlation.Active
		case hasInactive:
			componentNode.Status = correlation.Alarmed
		case hasUnknown:
			componentNode.Status = correlation.Unknown
		}

		if component.ParentIDs == nil {
			result = append(result, componentNode)
		}
	}

	return result
}
