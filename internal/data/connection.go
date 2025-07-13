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

func (m *ConnectionModel) Get(tenantID string) ([]*Connection, error) {
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
			group_concat(p2.port_network_component_id) as parent,
			group_concat(p3.port_network_component_id) as children
		FROM
			port p1
			LEFT OUTER JOIN network_component nc on nc.nc_id = p1.port_network_component_id
			LEFT OUTER JOIN port p2 on p2.port_id = p1.port_connected_to_port_id and p1.optical_signal_direction = 'RX'
			LEFT OUTER JOIN port p3 on p3.port_id = p1.port_connected_to_port_id and p1.optical_signal_direction = 'TX'
		WHERE
			p1.port_status = 'CONNECTED'
		GROUP BY
			p1.port_network_component_id
		HAVING
			parent IS NOT NULL or children IS NOT NULL;
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*Connection, 0)
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

		result = append(result, &connection)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
