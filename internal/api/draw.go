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

		// connections, err := models.Connection.GetAll(tenantID)
		// if err != nil {
		// 	serverErrorResponse(w, r, logger, err)
		// 	return
		// }

		// rootNodes := buildNetworkWithConnection(connections)

		segments, err := models.Segment.GetAll(tenantID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		// rootNodes := buildSegmentNodes(segments)

		components, err := models.Component.GetAll(tenantID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		rootNodes := buildComponentNodes(components, segments)

		if len(rootNodes) == 0 {
			serverErrorResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = core.Run(rootNode)
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

func buildNetworkWithConnection(connections []*data.Connection) []*data.Node {
	result := make([]*data.Node, 0)
	nodes := make(map[int]*data.Node)

	for _, connection := range connections {
		upsertConnectionMap(nodes, connection)

		if connection.ParentIDs == nil {
			result = append(result, nodes[connection.ID])
		}
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
	nodes[connection.ID] = data.NewNode(connection.ID, name, data.BoxNodeType)
}

func buildSegmentNodes(segments []*data.Segment) []*data.Node {
	result := make([]*data.Node, 0, len(segments))

	for _, segment := range segments {
		if segment.FiberIDs == nil {
			continue
		}

		name := fmt.Sprintf("%d - segment", segment.ID)
		segmentNode := data.NewNode(segment.ID, name, data.SegmentNodeType)

		fiberIDs := strings.SplitSeq(*segment.FiberIDs, ",")

		for fiberID := range fiberIDs {
			childID, err := strconv.Atoi(fiberID)
			if err != nil {
				continue
			}

			name := fmt.Sprintf("%d - fiber", childID)
			fiberNode := data.NewNode(childID, name, data.FiberNodeType)

			segmentNode.SetChildren(fiberNode)
		}

		result = append(result, segmentNode)
	}

	return result
}

func buildComponentNodes(components []*data.Component, segments []*data.Segment) []*data.Node {
	result := make([]*data.Node, 0)
	segmentNodes := make(map[int]*data.Node)

	for _, segment := range segments {
		name := fmt.Sprintf("%d - segment", segment.ID)
		segmentNodes[segment.ID] = data.NewNode(segment.ID, name, data.SegmentNodeType)
	}

	for _, component := range components {
		name := fmt.Sprintf("%d - component", component.ID)
		componentNode := data.NewNode(component.ID, name, data.BoxNodeType)

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
		}

		if component.ParentIDs == nil {
			result = append(result, componentNode)
		}
	}

	return result
}
