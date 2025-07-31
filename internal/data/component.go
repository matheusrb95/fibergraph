package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "embed"
)

//go:embed component.sql
var componentQuery string

type Component struct {
	ID       string
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
	rows, err := tx.QueryContext(ctx, componentQuery, projectID, projectID)
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
