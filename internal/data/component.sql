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
	p.port_splice_closure_network_component_id

UNION ALL

SELECT 
	f.fiber_segment_id,
	GROUP_CONCAT(f.fiber_id),
	'Segment'
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
