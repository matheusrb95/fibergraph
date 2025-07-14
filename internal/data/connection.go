package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Connection struct {
	ID          int
	Name        string
	ParentIDs   *string
	ChildrenIDs *string
}

type ConnectionModel struct {
	DB *sql.DB
}

func (m *ConnectionModel) GetAll(tenantID string) ([]*Connection, error) {
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

	connections, err := getConnections(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("get connection %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("commit %w", err)
	}

	return connections, nil
}

func getConnections(ctx context.Context, tx *sql.Tx) ([]*Connection, error) {
	query := `
		SELECT 
			p1.port_network_component_id,
			nc.nc_name,
			GROUP_CONCAT(p2.port_network_component_id) AS parent,
			GROUP_CONCAT(p3.port_network_component_id) AS children
		FROM
			port p1
			LEFT OUTER JOIN network_component nc ON nc.nc_id = p1.port_network_component_id
			LEFT OUTER JOIN port p2 ON p2.port_id = p1.port_connected_to_port_id AND p1.optical_signal_direction = 'RX'
			LEFT OUTER JOIN port p3 ON p3.port_id = p1.port_connected_to_port_id AND p1.optical_signal_direction = 'TX'
		WHERE
			p1.port_status = 'CONNECTED'
		GROUP BY
			p1.port_network_component_id
		HAVING
			parent IS NOT NULL OR children IS NOT NULL;
	`

	rows, err := tx.QueryContext(ctx, query)
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
