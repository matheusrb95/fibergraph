package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ONU struct {
	ID           string
	SerialNumber string
	Status       string
	FiberID      string
}

type ONUModel struct {
	DB *sql.DB
}

func (m *ONUModel) GetAll(tenantID, projectID string) ([]*ONU, error) {
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

	onus, err := getONUs(ctx, tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get onus %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return onus, nil
}

func getONUs(ctx context.Context, tx *sql.Tx, projectID string) ([]*ONU, error) {
	query := `
		SELECT
			o.onu_network_component_id,
			o.onu_gpon_serial_number,
			o.onu_operational_state,
			p2.port_network_component_id
		FROM
			onu o
			LEFT OUTER JOIN network_component nc ON nc.nc_id = o.onu_network_component_id
			LEFT OUTER JOIN project_network_component pnc ON pnc_network_component_id = nc.nc_id
			LEFT OUTER JOIN port p1 ON p1.port_network_component_id = nc.nc_id
			LEFT OUTER JOIN port p2 ON p2.port_connected_to_port_id = p1.port_id
		WHERE
			pnc.pnc_project_id = ?;
	`

	rows, err := tx.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	onus := make([]*ONU, 0)
	for rows.Next() {
		var onu ONU
		err := rows.Scan(
			&onu.ID,
			&onu.SerialNumber,
			&onu.Status,
			&onu.FiberID,
		)
		if err != nil {
			return nil, err
		}

		onus = append(onus, &onu)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return onus, nil
}
