package api

import (
	"log/slog"
	"net/http"

	"github.com/matheusrb95/fibergraph/internal/correlation"
	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/request"
	"github.com/matheusrb95/fibergraph/internal/response"
)

type EquipmentStatus struct {
	ActiveSensors  []string `json:"active_sensors"`
	AlarmedSensors []string `json:"alarmed_sensors"`
	ActiveONUs     []string `json:"active_onus"`
	AlarmedONUs    []string `json:"alarmed_onus"`
}

func HandleCorrelation(logger *slog.Logger, models *data.Models) http.Handler {
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

		var equipmentStatus EquipmentStatus
		err := request.DecodeJSON(w, r, &equipmentStatus)
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

		onus, err := models.ONU.GetAll(tenantID, projectID)
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
			"onus_len", len(onus),
			"segments_len", len(segments),
			"components_len", len(components),
		)

		c := correlation.New(
			connections,
			sensors,
			onus,
			equipmentStatus.ActiveSensors,
			equipmentStatus.AlarmedSensors,
			equipmentStatus.ActiveONUs,
			equipmentStatus.AlarmedONUs,
			segments,
			components,
		)
		if err := c.Run(); err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		err = response.JSON(w, http.StatusOK, response.Envelope{"correlation": "done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}
