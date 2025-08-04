package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/matheusrb95/fibergraph/internal/aws"
	"github.com/matheusrb95/fibergraph/internal/correlation"
	"github.com/matheusrb95/fibergraph/internal/data"
	"github.com/matheusrb95/fibergraph/internal/request"
	"github.com/matheusrb95/fibergraph/internal/response"
)

type EquipmentStatus struct {
	ActiveSensors   []string `json:"active_sensors"`
	AlarmedSensors  []string `json:"alarmed_sensors"`
	InactiveSensors []string `json:"inactive_sensors"`
	ActiveONUs      []string `json:"active_onus"`
	AlarmedONUs     []string `json:"alarmed_onus"`
}

func HandleCorrelation(logger *slog.Logger, models *data.Models, services *aws.Services) http.Handler {
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
			"components_len", len(components),
		)

		c := correlation.New(
			connections,
			sensors,
			onus,
			equipmentStatus.ActiveSensors,
			equipmentStatus.AlarmedSensors,
			equipmentStatus.InactiveSensors,
			equipmentStatus.ActiveONUs,
			equipmentStatus.AlarmedONUs,
			components,
		)
		if err := c.Run(); err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		projectIDint, err := strconv.Atoi(projectID)
		if err != nil {
			serverErrorResponse(w, r, logger, err)
			return
		}

		go func() {
			for _, node := range c.Result() {
				var topic string
				var msg *data.SNSMessage
				switch node.Type {
				case correlation.ONUNode:
					topic = "EH_ONU_EVENTS"
					msg = data.NewONUMessage(node.Type.String(), node.ID, node.Status.String(), tenantID, projectIDint, findOnuIDByNodeID(onus, node.ID))
				case correlation.SensorNode:
					topic = "EH_IOT_EVENTS"
					msg = data.NewSensorMessage(node.Type.String(), node.ID, node.Status.String(), tenantID, projectIDint)
				case correlation.FiberNode:
					continue
				default:
					topic = "EH_TOPOLOGIC_EVENTS"
					msg = data.NewSensorMessage(node.Type.String(), node.ID, node.Status.String(), tenantID, projectIDint)
				}

				jsonBytes, err := json.Marshal(msg)
				if err != nil {
					logger.Warn("error marshaling sns message", "err", err.Error())
					continue
				}

				err = services.SNS.Publish(string(jsonBytes), topic)
				if err != nil {
					logger.Warn("error sending sns message", "err", err.Error())
					continue
				}
				logger.Debug("sns message send.", "msg", string(jsonBytes))
			}
		}()

		err = response.JSON(w, http.StatusOK, response.Envelope{"correlation": "done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}

func findOnuIDByNodeID(onus []*data.ONU, nodeID string) string {
	for _, onu := range onus {
		if onu.SerialNumber != nodeID {
			continue
		}

		return onu.ID
	}

	return ""
}
