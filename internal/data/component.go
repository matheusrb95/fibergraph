package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Component struct {
	ID       int
	FiberIDs *string
}

type ComponentModel struct {
	DB *sql.DB
}

func (m *ComponentModel) GetAll(tenantID, projectID string) ([]*Component, error) {
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

	components, err := getComponents(ctx, tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get component %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return components, nil
}

func getComponents(ctx context.Context, tx *sql.Tx, projectID string) ([]*Component, error) {
	query := `
		SELECT 
			p.port_splice_closure_network_component_id,
			GROUP_CONCAT(p.port_network_component_id)
		FROM
			port p
			LEFT OUTER JOIN network_component nc ON nc.nc_id = p.port_splice_closure_network_component_id
			LEFT OUTER JOIN project_network_component pnc ON pnc.pnc_network_component_id = nc.nc_id
		WHERE
			p.optical_signal_direction = 'TX'
			AND pnc.pnc_project_id = ?
		GROUP BY
			p.port_splice_closure_network_component_id;
	`

	rows, err := tx.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	components := make([]*Component, 0)
	for rows.Next() {
		var component Component
		err := rows.Scan(
			&component.ID,
			&component.FiberIDs,
		)
		if err != nil {
			return nil, err
		}

		components = append(components, &component)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return components, nil
}
