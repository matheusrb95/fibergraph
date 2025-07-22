package core

import (
	"fmt"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"

	"github.com/matheusrb95/fibergraph/internal/correlation"
)

var counter int

func Run(node *correlation.Node, determine, draw, sensor bool) error {
	root := findRoot(node)
	if sensor {
		propagateSensorStatus(node)
	}

	if determine {
		determineNodeStatus(node)
	}

	if !draw {
		return nil
	}

	return drawGraphs(root)
}

func findRoot(node *correlation.Node) *correlation.Node {
	for node.Parents != nil {
		node = node.Parents[0]
	}

	return node
}

func propagateSensorStatus(node *correlation.Node) {
	if node.Children == nil {
		return
	}

	for _, child := range node.Children {
		propagateSensorStatus(child)

		if child.Type != correlation.SensorNode {
			continue
		}

		switch child.Status {
		case correlation.Active:
			activeAllAbove(node)
		case correlation.Alarmed:
			inactiveAllBelow(node)
		}
	}
}

func activeAllAbove(node *correlation.Node) {
	if node.Parents == nil {
		return
	}

	node.Status = correlation.Active

	for _, parent := range node.Parents {
		activeAllAbove(parent)
		parent.Status = correlation.Active
	}
}

func inactiveAllBelow(node *correlation.Node) {
	if node.Children == nil {
		return
	}

	node.Status = correlation.Alarmed

	for _, child := range node.Children {
		inactiveAllBelow(child)
		child.Status = correlation.Alarmed
	}
}

func determineNodeStatus(node *correlation.Node) {
	if node.Children == nil {
		return
	}

	var hasActive, hasInactive, hasUnknown bool

	for _, child := range node.Children {
		determineNodeStatus(child)

		switch child.Status {
		case correlation.Active:
			hasActive = true
		case correlation.Alarmed:
			hasInactive = true
		case correlation.Unknown:
			hasUnknown = true
		}
	}

	switch {
	case hasActive:
		node.Status = correlation.Active
	case hasInactive:
		node.Status = correlation.Alarmed
	case hasUnknown:
		node.Status = correlation.Unknown
	}
}

func drawGraphs(node *correlation.Node) error {
	g := graph.New(graph.IntHash, graph.Directed())

	err := walk(g, node)
	if err != nil {
		return err
	}

	counter++
	file, err := os.Create(fmt.Sprintf("%s-%d.gv", "my-graph", counter))
	if err != nil {
		return err
	}
	defer file.Close()

	err = draw.DOT(g, file)
	if err != nil {
		return err
	}

	return nil
}

func walk(g graph.Graph[int, int], n *correlation.Node) error {
	var attr func(*graph.VertexProperties)
	attr = graph.VertexAttribute("color", "black")

	switch n.Status {
	case correlation.Alarmed:
		attr = graph.VertexAttribute("color", "red")
	case correlation.Active:
		attr = graph.VertexAttribute("color", "green")
	default:
		attr = graph.VertexAttribute("color", "black")
	}
	_ = g.AddVertex(n.ID, graph.VertexAttribute("label", n.Name), attr)

	for _, child := range n.Children {
		var attr func(*graph.VertexProperties)

		switch child.Status {
		case correlation.Alarmed:
			attr = graph.VertexAttribute("color", "red")
		case correlation.Active:
			attr = graph.VertexAttribute("color", "green")
		default:
			attr = graph.VertexAttribute("color", "black")
		}

		err := g.AddVertex(child.ID, graph.VertexAttribute("label", child.Name), attr)
		if err != nil {
			return fmt.Errorf("add vertex. %w", err)
		}
		err = g.AddEdge(n.ID, child.ID)
		if err != nil {
			return fmt.Errorf("add edge. %w", err)
		}

		err = walk(g, child)
		if err != nil {
			return fmt.Errorf("walk. %w", err)
		}
	}

	return nil
}
