package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type SNSMessage struct {
	Timestamp            time.Time `json:"timestamp"`
	NetworkComponentType string    `json:"network_component_type"`
	NetworkComponentID   string    `json:"network_component_id"`
	Description          string    `json:"description"`
	Status               string    `json:"status"`
	AlarmedProbability   string    `json:"alarmedProbability"`
	AlarmedBox           int       `json:"alarmed_box"`
	Last                 bool      `json:"last"`
	RootID               int       `json:"rootID"`
	ProjectID            int       `json:"projectID"`
	TenantID             string    `json:"tenant_id"`
	DevEUI               string    `json:"dev_eui,omitempty"`
	Cause                []string  `json:"cause,omitempty"`
	OpticalPower         float32   `json:"opticalPower,omitempty"`
	SerialNumber         string    `json:"serial_number"`
	ONUSerialNumber      string    `json:"onu_sn,omitempty"`
	ONUID                string    `json:"onu_id,omitempty"`
	ONUMessage           string    `json:"message,omitempty"`
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
			equipmentStatus.InactiveSensors,
			equipmentStatus.ActiveONUs,
			equipmentStatus.AlarmedONUs,
			segments,
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
				networkComponentType := strings.ToUpper(node.Type.String())
				status := strings.ToUpper(node.Status.String())

				var devEUI string
				var opticalPower float32
				var cause []string
				if node.Type == correlation.SensorNode {
					for _, sensor := range sensors {
						if sensor.ID+1_000_000 != node.ID {
							continue
						}

						devEUI = sensor.DevEUI
						opticalPower = -10.0
						cause = []string{"OPTICAL_POWER_ALERT"}
						break
					}
				}

				var onuSerialNumber, onuID, onuMessage string
				if node.Type == correlation.ONUNode {
					for _, onu := range onus {
						if onu.ID != node.ID {
							continue
						}

						onuSerialNumber = onu.Serial
						onuID = strconv.Itoa(node.ID)
						switch node.Status {
						case correlation.Active:
							onuMessage = "ONU activated"
						case correlation.Alarmed:
							onuMessage = "ONU deactivated"
						}
						break
					}
				}

				var alarmedProbability string
				var alarmedBox int
				switch node.Status {
				case correlation.Alarmed:
					alarmedProbability = "1.0"
					alarmedBox = 1
				default:
					alarmedProbability = "0.0"
					alarmedBox = 0
				}

				msg := SNSMessage{
					Timestamp:            time.Now(),
					NetworkComponentType: networkComponentType,
					NetworkComponentID:   strconv.Itoa(node.ID),
					Description:          fmt.Sprintf("%s %s", status, networkComponentType),
					Status:               status,
					AlarmedProbability:   alarmedProbability,
					AlarmedBox:           alarmedBox,
					Last:                 false,
					RootID:               0,
					ProjectID:            projectIDint,
					TenantID:             tenantID,
					DevEUI:               devEUI,
					Cause:                cause,
					OpticalPower:         opticalPower,
					ONUSerialNumber:      onuSerialNumber,
					ONUID:                onuID,
					ONUMessage:           onuMessage,
				}

				jsonBytes, err := json.Marshal(msg)
				if err != nil {
					logger.Warn("error marshaling sns message", "err", err.Error())
					continue
				}

				var topic string
				switch node.Type {
				case correlation.ONUNode:
					topic = "EH_ONU_EVENTS"
				case correlation.SensorNode:
					topic = "EH_IOT_EVENTS"
				case correlation.FiberNode:
					continue
				default:
					topic = "EH_TOPOLOGIC_EVENTS"
				}

				err = services.SNS.Publish(string(jsonBytes), topic)
				if err != nil {
					logger.Warn("error sending sns message", "err", err.Error())
					continue
				}
				logger.Info("sns message send.", "msg", string(jsonBytes))
			}
		}()

		err = response.JSON(w, http.StatusOK, response.Envelope{"correlation": "done"})
		if err != nil {
			serverErrorResponse(w, r, logger, err)
		}
	})
}
