package main

import (
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

var id int

func ID() int {
	id++
	return id
}

func main() {
	node := network2()
	inactivateNodes(node)
	findRootCauses(node)
	drawGraphs(node)
}

func findRoot(node *Node) *Node {
	for node.parent != nil {
		node = node.parent
	}

	return node
}

func inactivateNodes(node *Node) {
	for _, child := range node.children {
		inactivateNodes(child)

		if Inactive(child) {
			child.Active = false
		}
	}
}

func Inactive(node *Node) bool {
	if node.children == nil {
		return false
	}

	var inactives int
	for _, child := range node.children {
		if child.Active {
			continue
		}

		inactives++
	}

	return inactives == len(node.children)
}

func findRootCauses(node *Node) {
	if Inactive(node) {
		node.RootCause = true
		return
	}

	for _, child := range node.children {
		if Inactive(node) {
			node.RootCause = true
			continue
		}

		findRootCauses(child)
	}
}

func drawGraphs(node *Node) {
	g := graph.New(graph.IntHash, graph.Directed())

	root := findRoot(node)

	var walk func(n *Node)
	walk = func(n *Node) {
		var attr func(*graph.VertexProperties)
		if node.RootCause {
			attr = graph.VertexAttribute("color", "blue")
		} else {
			attr = graph.VertexAttribute("color", "black")
		}
		_ = g.AddVertex(n.ID, graph.VertexAttribute("label", n.Name), attr)

		for _, child := range n.children {
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
