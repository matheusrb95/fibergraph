package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "embed"
)

//go:embed connection.sql
var connectionQuery string

type Connection struct {
	ID          string
	Name        string
	ParentIDs   *string
	ChildrenIDs *string
	Type        string
}

type ConnectionModel struct {
	DB *sql.DB
}

func (m *ConnectionModel) GetAll(tenantID, projectID string) ([]*Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	connections, err := getConnections(ctx, tx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get connection %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return connections, nil
}

func getConnections(ctx context.Context, tx *sql.Tx, projectID string) ([]*Connection, error) {
	args := []any{projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID}
	rows, err := tx.QueryContext(ctx, connectionQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	connections := make([]*Connection, 0)
	for rows.Next() {
		var connection Connection
		err := rows.Scan(
			&connection.ID,
			&connection.Name,
			&connection.ParentIDs,
			&connection.ChildrenIDs,
			&connection.Type,
		)
		if err != nil {
			return nil, err
		}

		connections = append(connections, &connection)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}
