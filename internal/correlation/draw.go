package correlation

import (
	"fmt"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

var counter int

func drawGraphs(node *Node) error {
	g := graph.New(graph.IntHash, graph.Directed())

	walk(g, node)

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

func walk(g graph.Graph[int, int], n *Node) {
	var attr func(*graph.VertexProperties)
	attr = graph.VertexAttribute("color", "black")

	switch n.Status {
	case Alarmed:
		attr = graph.VertexAttribute("color", "red")
	case ProbablyAlarmed:
		attr = graph.VertexAttribute("color", "orange")
	case Inconsistent:
		attr = graph.VertexAttribute("color", "pink")
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
		case ProbablyAlarmed:
			attr = graph.VertexAttribute("color", "orange")
		case Active:
			attr = graph.VertexAttribute("color", "green")
		case Inconsistent:
			attr = graph.VertexAttribute("color", "pink")
		default:
			attr = graph.VertexAttribute("color", "black")
		}

		_ = g.AddVertex(child.ID, graph.VertexAttribute("label", child.Name), attr)
		_ = g.AddEdge(n.ID, child.ID)

		walk(g, child)
	}
}
