package cmd

import (
	"sheeva/config"
	"strings"
)

type Node struct {
	Group   *config.GitlabElement `json:"group"`
	Nodes   map[int][]*Node       `json:"nodes,omitempty"`
	visited bool
	seq     int
	edge    int
}

func (n *Node) AddNode(group config.GitlabElement, edge int) {
	n.Nodes[edge] = append(n.Nodes[edge], NewNode(group, edge))
}

func (n *Node) GetEdgeNodes(edge int) []*Node {
	if v, ok := n.Nodes[edge]; ok {
		return v
	}
	return nil
}

func NewNode(root config.GitlabElement, edge int) *Node {
	return &Node{
		Group: &root,
		Nodes: make(map[int][]*Node),
		edge:  edge,
	}
}

func NewGroupGraphs(groups []config.GitlabElement) []*Node {
	var graphs []*Node

	// Create roots
	for _, g := range groups {
		if g.Namespace == g.Name {
			gg := NewNode(g, 0)
			graphs = append(graphs, gg)
		}
	}

	// Create edges
	for _, root := range graphs {
		for _, g := range groups {
			if strings.Contains(g.Namespace, root.Group.Name) {
				root.AddNode(g, len(strings.Split(g.Namespace, "/")))
			}
		}
	}

	return graphs
}
