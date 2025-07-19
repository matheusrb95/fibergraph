package core

import (
	"fmt"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"

	"github.com/matheusrb95/fibergraph/internal/data"
)

var counter int

func Run(node *data.Node, determine, draw, sensor bool) error {
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

func findRoot(node *data.Node) *data.Node {
	for node.Parent != nil {
		node = node.Parent
	}

	return node
}

func propagateSensorStatus(node *data.Node) {
	if node.Children == nil {
		return
	}

	for _, child := range node.Children {
		propagateSensorStatus(child)

		if child.Type != data.SensorNode {
			continue
		}

		switch child.Status {
		case data.Active:
			activeAllAbove(child)
		case data.Inactive:
			inactiveAllBelow(node)
		}
	}
}

func activeAllAbove(node *data.Node) {
	for node.Parent != nil {
		node.Status = data.Active
		node = node.Parent
	}
}

func inactiveAllBelow(node *data.Node) {
	if node.Children == nil {
		return
	}

	node.Status = data.Inactive

	for _, child := range node.Children {
		inactiveAllBelow(child)
		child.Status = data.Inactive
	}
}

func determineNodeStatus(node *data.Node) {
	if node.Children == nil {
		return
	}

	var hasActive, hasInactive, hasUnknown bool

	for _, child := range node.Children {
		determineNodeStatus(child)

		switch child.Status {
		case data.Active:
			hasActive = true
		case data.Inactive:
			hasInactive = true
		case data.Unknown:
			hasUnknown = true
		}
	}

	switch {
	case hasActive:
		node.Status = data.Active
	case hasInactive:
		node.Status = data.Inactive
	case hasUnknown:
		node.Status = data.Unknown
	}
}

func drawGraphs(node *data.Node) error {
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

func walk(g graph.Graph[int, int], n *data.Node) error {
	var attr func(*graph.VertexProperties)
	if n.RootCause {
		attr = graph.VertexAttribute("color", "blue")
	} else {
		attr = graph.VertexAttribute("color", "black")
	}

	switch n.Status {
	case data.Inactive:
		attr = graph.VertexAttribute("color", "red")
	case data.Active:
		attr = graph.VertexAttribute("color", "green")
	default:
		attr = graph.VertexAttribute("color", "black")
	}
	_ = g.AddVertex(n.ID, graph.VertexAttribute("label", n.Name), attr)

	for _, child := range n.Children {
		var attr func(*graph.VertexProperties)

		switch child.Status {
		case data.Inactive:
			attr = graph.VertexAttribute("color", "red")
		case data.Active:
			attr = graph.VertexAttribute("color", "green")
		default:
			attr = graph.VertexAttribute("color", "black")
		}

		if child.RootCause {
			attr = graph.VertexAttribute("color", "blue")
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
