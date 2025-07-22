package correlation

import (
	"fmt"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

var counter int

func Run(node *Node, draw, sensor bool) error {
	root := findRoot(node)
	if sensor {
		propagateSensorStatus(node)
	}

	if !draw {
		return nil
	}

	return drawGraphs(root)
}

func findRoot(node *Node) *Node {
	for node.Parents != nil {
		node = node.Parents[0]
	}

	return node
}

func propagateSensorStatus(node *Node) {
	if node.Children == nil {
		return
	}

	for _, child := range node.Children {
		propagateSensorStatus(child)

		if child.Type != SensorNode {
			continue
		}

		switch child.Status {
		case Active:
			activeAllAbove(node)
		case Alarmed:
			inactiveAllBelow(node)
		}
	}
}

func activeAllAbove(node *Node) {
	if node.Parents == nil {
		return
	}

	node.Status = Active

	for _, parent := range node.Parents {
		activeAllAbove(parent)
		parent.Status = Active
	}
}

func inactiveAllBelow(node *Node) {
	if node.Children == nil {
		return
	}

	node.Status = Alarmed

	for _, child := range node.Children {
		inactiveAllBelow(child)
		child.Status = Alarmed
	}
}

func drawGraphs(node *Node) error {
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

func walk(g graph.Graph[int, int], n *Node) error {
	var attr func(*graph.VertexProperties)
	attr = graph.VertexAttribute("color", "black")

	switch n.Status {
	case Alarmed:
		attr = graph.VertexAttribute("color", "red")
	case Active:
		attr = graph.VertexAttribute("color", "green")
	default:
		attr = graph.VertexAttribute("color", "black")
	}
	_ = g.AddVertex(n.ID, graph.VertexAttribute("label", n.Name), attr)

	for _, child := range n.Children {
		var attr func(*graph.VertexProperties)

		switch child.Status {
		case Alarmed:
			attr = graph.VertexAttribute("color", "red")
		case Active:
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
