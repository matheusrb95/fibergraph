package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Segment struct {
	ID       int
	FiberIDs *string
}

type SegmentModel struct {
	DB *sql.DB
}

func (m *SegmentModel) GetAll(tenantID string) ([]*Segment, error) {
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

	segments, err := getSegments(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("get segment %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return segments, nil
}

func getSegments(ctx context.Context, tx *sql.Tx) ([]*Segment, error) {
	query := `
		SELECT 
			f.fiber_segment_id,
			GROUP_CONCAT(f.fiber_id)
		FROM
			fiber f
		GROUP BY
			f.fiber_segment_id;
	`

	rows, err := tx.QueryContext(ctx, query)
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
