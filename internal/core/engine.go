package core

import (
	"fmt"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"

	"github.com/matheusrb95/fibergraph/internal/data"
)

func Run(node *data.Node) error {
	//inactivateNodes(node)
	//findRootCauses(node)
	return drawGraphs(node)
}

func findRoot(node *data.Node) *data.Node {
	for node.Parent != nil {
		node = node.Parent
	}

	return node
}

func inactivateNodes(node *data.Node) {
	for _, child := range node.Children {
		inactivateNodes(child)

		if inactive(child) {
			child.Active = false
		}
	}
}

func inactive(node *data.Node) bool {
	if node.Children == nil {
		return false
	}

	var inactives int
	for _, child := range node.Children {
		if child.Active {
			continue
		}

		inactives++
	}

	return inactives == len(node.Children)
}

func findRootCauses(node *data.Node) {
	if inactive(node) {
		node.RootCause = true
		return
	}

	for _, child := range node.Children {
		if inactive(node) {
			node.RootCause = true
			continue
		}

		findRootCauses(child)
	}
}

func drawGraphs(node *data.Node) error {
	g := graph.New(graph.IntHash, graph.Directed())

	root := findRoot(node)

	var walk func(n *data.Node) error
	walk = func(n *data.Node) error {
		var attr func(*graph.VertexProperties)
		if node.RootCause {
			attr = graph.VertexAttribute("color", "blue")
		} else {
			attr = graph.VertexAttribute("color", "black")
		}

		_ = g.AddVertex(n.ID, graph.VertexAttribute("label", n.Name), attr)

		for _, child := range n.Children {
			var attr func(*graph.VertexProperties)
			if !child.Active {
				attr = graph.VertexAttribute("color", "red")
			} else {
				attr = graph.VertexAttribute("color", "green")
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

			err = walk(child)
			if err != nil {
				return fmt.Errorf("walk. %w", err)
			}
		}

		return nil
	}

	err := walk(root)
	if err != nil {
		return err
	}

	file, err := os.Create("my-graph.gv")
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
