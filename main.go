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

func network1() *Node {
	Box1 := NewNode(ID(), "Box1", BoxNodeType)
	Box2 := NewNode(ID(), "Box2", BoxNodeType)
	Box3 := NewNode(ID(), "Box3", BoxNodeType)

	ONU1 := NewNode(ID(), "ONU1", ONUNodeType)
	ONU1.Active = false

	ONU2 := NewNode(ID(), "ONU2", ONUNodeType)
	ONU2.Active = false

	ONU3 := NewNode(ID(), "ONU3", ONUNodeType)
	ONU3.Active = false

	ONU4 := NewNode(ID(), "ONU4", ONUNodeType)
	//ONU4.Active = false

	OLT := NewNode(ID(), "OLT", OLTNodeType)

	Splitter1 := NewNode(ID(), "Splitter1", SplitterNodeType)
	Splitter2 := NewNode(ID(), "Splitter2", SplitterNodeType)
	Splitter3 := NewNode(ID(), "Splitter3", SplitterNodeType)

	Box1.SetParent(OLT)
	Box1.SetChildren(Splitter1)
	Splitter1.SetChildren(ONU1, ONU2)

	Box2.SetParent(OLT)
	Box2.SetChildren(Splitter2)
	Splitter2.SetChildren(ONU3)

	Box3.SetParent(OLT)
	Box3.SetChildren(Splitter3)
	Splitter3.SetChildren(ONU4)

	ONU1.SetParent(Splitter1)
	ONU2.SetParent(Splitter1)
	ONU3.SetParent(Splitter2)
	ONU4.SetParent(Splitter3)

	OLT.SetChildren(Box1, Box2, Box3)

	return OLT
}

func main() {
	node := network1()
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
				attr = graph.VertexAttribute("color", "black")
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
