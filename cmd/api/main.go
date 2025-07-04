package main

import (
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"

	"github.com/matheusrb95/fibergraph/internal/data"
)

func main() {
	node := data.Network3()
	inactivateNodes(node)
	findRootCauses(node)
	drawGraphs(node)
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

		if Inactive(child) {
			child.Active = false
		}
	}
}

func Inactive(node *data.Node) bool {
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
	if Inactive(node) {
		node.RootCause = true
		return
	}

	for _, child := range node.Children {
		if Inactive(node) {
			node.RootCause = true
			continue
		}

		findRootCauses(child)
	}
}

func drawGraphs(node *data.Node) {
	g := graph.New(graph.IntHash, graph.Directed())

	root := findRoot(node)

	var walk func(n *data.Node)
	walk = func(n *data.Node) {
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

			_ = g.AddVertex(child.ID, graph.VertexAttribute("label", child.Name), attr)
			_ = g.AddEdge(n.ID, child.ID)

			walk(child)
		}
	}

	walk(root)

	file, _ := os.Create("my-graph.gv")
	defer file.Close()
	_ = draw.DOT(g, file)
}
