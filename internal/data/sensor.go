package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Sensor struct {
	ID      int
	DevEUI  string
	Status  string
	FiberID int
}

type SensorModel struct {
	DB *sql.DB
}

func (m *SensorModel) GetAll(tenantID string) ([]*Sensor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin %w", err)
	}
	defer tx.Rollback()

	err = setSchema(ctx, tx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("set schema %w", err)
	}

	sensors, err := getSensors(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("get sensor %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return sensors, nil
}

func getSensors(ctx context.Context, tx *sql.Tx) ([]*Sensor, error) {
	query := `
		SELECT 
			s.sensor_id,
			s.sensor_deveui,
			s.sensor_operational_status,
			p.port_network_component_id
		FROM
			sensor s
			LEFT OUTER JOIN port p ON p.port_id = s.sensor_port_id;
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sensors := make([]*Sensor, 0)
	for rows.Next() {
		var sensor Sensor
		err := rows.Scan(
			&sensor.ID,
			&sensor.DevEUI,
			&sensor.Status,
			&sensor.FiberID,
		)
		if err != nil {
			return nil, err
		}

		sensors = append(sensors, &sensor)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sensors, nil
}
