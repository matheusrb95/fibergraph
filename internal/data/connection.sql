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
	AND o.onu_network_component_id IS NULL
    AND cto2.cto_network_component_id IS NULL
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
	p1.port_network_component_id AS port_network_component_id,
    'no_name',
	p2.port_network_component_id AS parent,
    null,
    'Fiber'
FROM
	cto_connection cc 
	LEFT OUTER JOIN port src ON src.port_id = cc.ctc_port_out_id
	LEFT OUTER JOIN port dst ON dst.port_id = cc.ctc_port_in_id
	
	JOIN port p1 ON p1.port_id = src.port_connected_to_port_id
	JOIN port p2 ON p2.port_id = dst.port_connected_to_port_id

    LEFT OUTER JOIN cto ON cto.cto_network_component_id = cc.ctc_cto_id
    LEFT OUTER JOIN network_component nc ON nc.nc_id = cto.cto_network_component_id
    LEFT OUTER JOIN project_network_component pnc ON pnc.pnc_network_component_id = nc.nc_id
WHERE
	pnc.pnc_project_id = ?

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
