package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Segment struct {
	ID       string
	FiberIDs *string
}

type SegmentModel struct {
	DB *sql.DB
}

func (m *SegmentModel) GetAll(tenantID, projectID string) ([]*Segment, error) {
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

	segments, err := getSegments(ctx, tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get segment %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return segments, nil
}

func getSegments(ctx context.Context, tx *sql.Tx, projectID string) ([]*Segment, error) {
	query := `
		SELECT 
			f.fiber_segment_id,
			GROUP_CONCAT(f.fiber_id)
		FROM
			fiber f
			LEFT OUTER JOIN segment s ON s.segment_id = f.fiber_segment_id
			LEFT OUTER JOIN cable c ON c.cable_id = s.segment_cable_id
			LEFT OUTER JOIN network_component nc ON nc.nc_id = c.cable_id
			LEFT OUTER JOIN project_network_component pnc ON pnc_network_component_id = nc.nc_id
		WHERE
			pnc.pnc_project_id = ?
		GROUP BY
			f.fiber_segment_id;
	`

	rows, err := tx.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	segments := make([]*Segment, 0)
	for rows.Next() {
		var segment Segment
		err := rows.Scan(
			&segment.ID,
			&segment.FiberIDs,
		)
		if err != nil {
			return nil, err
		}

		segments = append(segments, &segment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return segments, nil
}
