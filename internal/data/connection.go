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
	query := `
		SELECT 
			p1.port_network_component_id,
			nc.nc_name,
			CASE
				WHEN d.dio_network_component_id IS NOT NULL THEN d.dio_co_network_component_id ELSE GROUP_CONCAT(p2.port_network_component_id)
			END AS parent,
			GROUP_CONCAT(p3.port_network_component_id) AS children,
			CASE
				WHEN f.fiber_id IS NOT NULL THEN 'Fiber'
				WHEN s.splitter_network_component_id IS NOT NULL THEN 'Splitter'
				WHEN d.dio_network_component_id IS NOT NULL THEN 'DIO'
				WHEN o.onu_network_component_id IS NOT NULL THEN 'ONU'
				WHEN cto2.cto_network_component_id IS NOT NULL THEN 'CTO'
			END
		FROM
			port p1
			LEFT OUTER JOIN network_component nc ON nc.nc_id = p1.port_network_component_id
			
			LEFT OUTER JOIN fiber f ON f.fiber_id = p1.port_network_component_id
			LEFT OUTER JOIN segment sg ON sg.segment_id = f.fiber_segment_id
			LEFT OUTER JOIN cable c ON c.cable_id = sg.segment_cable_id
			LEFT OUTER JOIN network_component nc1 ON nc1.nc_id = c.cable_id
			LEFT OUTER JOIN project_network_component pnc1 ON pnc1.pnc_network_component_id = nc1.nc_id

			LEFT OUTER JOIN splitter s ON s.splitter_network_component_id = p1.port_network_component_id
			LEFT OUTER JOIN cto_splitter cs1 ON cs1.cto_splitter_splitter_id = s.splitter_network_component_id
			LEFT OUTER JOIN cto ON cto.cto_network_component_id = cs1.cto_splitter_cto_id
			LEFT OUTER JOIN network_component nc2 ON nc2.nc_id = cto.cto_network_component_id
			LEFT OUTER JOIN project_network_component pnc2 ON pnc2.pnc_network_component_id = nc2.nc_id
			
			LEFT OUTER JOIN ceo_splitter cs2 ON cs2.ceo_splitter_splitter_id = s.splitter_network_component_id
			LEFT OUTER JOIN ceo ON ceo.ceo_network_component_id = cs2.ceo_splitter_ceo_id
			LEFT OUTER JOIN network_component nc3 ON nc3.nc_id = ceo.ceo_network_component_id
			LEFT OUTER JOIN project_network_component pnc3 ON pnc3.pnc_network_component_id = nc3.nc_id

			LEFT OUTER JOIN dio d ON d.dio_network_component_id = p1.port_network_component_id
			LEFT OUTER JOIN network_component nc4 ON nc4.nc_id = d.dio_network_component_id
			LEFT OUTER JOIN project_network_component pnc4 ON pnc4.pnc_network_component_id = nc4.nc_id

			LEFT OUTER JOIN onu o ON o.onu_network_component_id = p1.port_network_component_id
			LEFT OUTER JOIN network_component nc5 ON nc5.nc_id = o.onu_network_component_id
			LEFT OUTER JOIN project_network_component pnc5 ON pnc5.pnc_network_component_id = nc5.nc_id

			LEFT OUTER JOIN cto cto2 ON cto2.cto_network_component_id = p1.port_network_component_id
			LEFT OUTER JOIN network_component nc6 ON nc6.nc_id = cto2.cto_network_component_id
			LEFT OUTER JOIN project_network_component pnc6 ON pnc6.pnc_network_component_id = nc6.nc_id

			LEFT OUTER JOIN port p2 ON p2.port_id = p1.port_connected_to_port_id AND p1.optical_signal_direction = 'RX'
			LEFT OUTER JOIN port p3 ON p3.port_id = p1.port_connected_to_port_id AND p1.optical_signal_direction = 'TX'
		WHERE
			p1.port_status = 'CONNECTED'
			AND (
				pnc1.pnc_project_id = ?
				OR pnc2.pnc_project_id = ?
				OR pnc3.pnc_project_id = ?
				OR pnc4.pnc_project_id = ?
				OR pnc5.pnc_project_id = ?
				OR pnc6.pnc_project_id = ?
			)
		GROUP BY
			p1.port_network_component_id
		HAVING
			parent IS NOT NULL OR children IS NOT NULL
			
		UNION ALL

		SELECT
			c.co_network_component_id,
			nc.nc_name,
			null,
			group_concat(d.dio_network_component_id),
			'CO'
		FROM
			co c
			LEFT OUTER JOIN network_component nc ON nc.nc_id = c.co_network_component_id
			LEFT OUTER JOIN dio d ON d.dio_co_network_component_id = c.co_network_component_id
			LEFT OUTER JOIN project_network_component pnc ON pnc.pnc_network_component_id = nc.nc_id
		WHERE
			pnc.pnc_project_id = ?
		GROUP BY 
			c.co_network_component_id;
	`

	args := []any{projectID, projectID, projectID, projectID, projectID, projectID, projectID}
	rows, err := tx.QueryContext(ctx, query, args...)
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
