package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Component struct {
	ID          int
	Name        string
	ParentIDs   *string
	ChildrenIDs *string
}

type ComponentModel struct {
	DB *sql.DB
}

func (m *ComponentModel) GetAll(tenantID string) ([]*Component, error) {
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

	components, err := getComponents(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("get component %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return components, nil
}

func getComponents(ctx context.Context, tx *sql.Tx) ([]*Component, error) {
	query := `
		SELECT 
			csc.csc_network_component_id,
			nc.nc_name,
			GROUP_CONCAT(CASE WHEN csc.csc_segment_point = 'END' THEN csc.csc_segment_id ELSE NULL END),
			GROUP_CONCAT(CASE WHEN csc.csc_segment_point = 'START' THEN csc.csc_segment_id ELSE NULL END)
		FROM
			component_segment_connection csc
			LEFT OUTER JOIN network_component nc ON nc.nc_id = csc.csc_network_component_id
		WHERE
			csc.csc_network_component_id IS NOT NULL
		GROUP BY
			csc.csc_network_component_id, nc.nc_name;
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	components := make([]*Component, 0)
	for rows.Next() {
		var component Component
		err := rows.Scan(
			&component.ID,
			&component.Name,
			&component.ParentIDs,
			&component.ChildrenIDs,
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
