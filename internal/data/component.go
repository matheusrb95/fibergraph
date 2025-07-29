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
	Type     string
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
			GROUP_CONCAT(p.port_network_component_id),
			CASE
				WHEN ceo.ceo_network_component_id IS NOT NULL THEN 'CEO'
				WHEN cto.cto_network_component_id IS NOT NULL THEN 'CTO'
				WHEN co.co_network_component_id IS NOT NULL THEN 'CO'
				WHEN onu.onu_network_component_id IS NOT NULL THEN 'ONU'
			END
		FROM
			port p
			LEFT OUTER JOIN network_component nc ON nc.nc_id = p.port_splice_closure_network_component_id
			LEFT OUTER JOIN project_network_component pnc ON pnc.pnc_network_component_id = nc.nc_id

			LEFT OUTER JOIN ceo ON ceo.ceo_network_component_id = nc.nc_id
			LEFT OUTER JOIN cto ON cto.cto_network_component_id = nc.nc_id
			LEFT OUTER JOIN co ON co.co_network_component_id = nc.nc_id
			LEFT OUTER JOIN onu ON onu.onu_network_component_id = nc.nc_id
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
			&component.Type,
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
