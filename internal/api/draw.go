package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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

		err = core.Run(buildNetwork(connections))
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		err = response.JSON(w, http.StatusCreated, response.Envelope{"message": "draw done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}

func buildNetwork(connections []*data.Connection) *data.Node {
	nodes := make(map[int]*data.Node)

	for _, connection := range connections {
		upsertMap(nodes, connection)
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

	for _, v := range nodes {
		return v
	}

	return nil
}

func upsertMap(nodes map[int]*data.Node, connection *data.Connection) {
	if _, ok := nodes[connection.ID]; ok {
		return
	}

	name := fmt.Sprintf("%d - %s", connection.ID, connection.Name)
	nodes[connection.ID] = data.NewNode(connection.ID, name, data.BoxNodeType)
}
