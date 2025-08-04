package data

import (
	"fmt"
	"time"
)

type SNSMessage struct {
	Timestamp            time.Time `json:"timestamp"`
	NetworkComponentType string    `json:"network_component_type"`
	NetworkComponentID   string    `json:"network_component_id"`
	Description          string    `json:"description"`
	Status               string    `json:"status"`
	AlarmedProbability   string    `json:"alarmedProbability"`
	AlarmedBox           int       `json:"alarmed_box,omitempty"`
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

func NewComponentMessage(ncType, ncID, status, tenantID string, projectID int) *SNSMessage {
	var alarmedProbability string
	var alarmedBox int
	switch status {
	case "ALARMED":
		alarmedProbability = "1.0"
		alarmedBox = 1
	default:
		alarmedProbability = "0.0"
		alarmedBox = 0
	}

	return &SNSMessage{
		Timestamp:            time.Now(),
		NetworkComponentType: ncType,
		NetworkComponentID:   ncID,
		Description:          fmt.Sprintf("%s %s", status, ncType),
		Status:               status,
		AlarmedProbability:   alarmedProbability,
		AlarmedBox:           alarmedBox,
		ProjectID:            projectID,
		TenantID:             tenantID,
	}
}

func NewSensorMessage(ncType, ncID, status, tenantID string, projectID int) *SNSMessage {
	var alarmedProbability string
	switch status {
	case "ALARMED":
		alarmedProbability = "1.0"
	default:
		alarmedProbability = "0.0"
	}

	return &SNSMessage{
		Timestamp:            time.Now(),
		NetworkComponentType: ncType,
		NetworkComponentID:   ncID,
		Description:          fmt.Sprintf("%s %s", status, ncType),
		Status:               status,
		AlarmedProbability:   alarmedProbability,
		ProjectID:            projectID,
		TenantID:             tenantID,
		DevEUI:               ncID,
		Cause:                []string{"OPTICAL_POWER_ALERT"},
		OpticalPower:         -10.0,
	}
}

func NewONUMessage(ncType, ncID, status, tenantID string, projectID int, onuID string) *SNSMessage {
	var onuMessage string
	switch status {
	case "ALARMED":
		onuMessage = "ONU deactivated"
	default:
		onuMessage = "ONU activated"
	}

	return &SNSMessage{
		Timestamp:            time.Now(),
		NetworkComponentType: ncType,
		NetworkComponentID:   ncID,
		Description:          fmt.Sprintf("%s %s", status, ncType),
		Status:               status,
		ProjectID:            projectID,
		TenantID:             tenantID,
		ONUSerialNumber:      ncID,
		ONUID:                onuID,
		ONUMessage:           onuMessage,
	}
}
