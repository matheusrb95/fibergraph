package api

import (
	"errors"
	"log/slog"
	"net/http"

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

		segments, err := models.Segment.GetAll(tenantID, projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		components, err := models.Component.GetAll(tenantID, projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		logger.Info("network size",
			"tenant_id", tenantID,
			"project_id", projectID,
			"connections_len", len(connections),
			"sensors_len", len(sensors),
			"segments_len", len(segments),
			"components_len", len(components),
		)

		nodes := make(map[int]*correlation.Node)

		rootNodes := correlation.BuildNetworkWithConnection(nodes, connections, sensors, sensorStatus.Active, sensorStatus.Alarmed)

		if len(rootNodes) == 0 {
			badRequestResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = correlation.Run(rootNode, true, true)
			if err != nil {
				serverErrorResponse(w, r, logger, err)
				return
			}
			break
		}

		segmentNodes := correlation.BuildSegmentNodes(nodes, segments)

		rootNodes = correlation.BuildComponentNodes(components, segmentNodes)

		if len(rootNodes) == 0 {
			badRequestResponse(w, r, logger, errors.New("no nodes"))
			return
		}

		for _, rootNode := range rootNodes {
			err = correlation.Run(rootNode, true, false)
			if err != nil {
				serverErrorResponse(w, r, logger, err)
				return
			}
			break
		}

		err = response.JSON(w, http.StatusCreated, response.Envelope{"message": "draw done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}
