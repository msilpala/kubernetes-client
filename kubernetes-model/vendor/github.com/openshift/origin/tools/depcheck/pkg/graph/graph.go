/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package graph

import (
	"fmt"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/encoding/dot"
)

type Node struct {
	Id         int
	UniqueName string
	LabelName  string
	Color      string
}

func (n Node) ID() int {
	return n.Id
}

func (n Node) String() string {
	return n.LabelName
}

// DOTAttributes implements an attribute getter for the DOT encoding
func (n Node) DOTAttributes() []dot.Attribute {
	color := n.Color
	if len(color) == 0 {
		color = "black"
	}

	return []dot.Attribute{
		{Key: "label", Value: fmt.Sprintf("%q", n.LabelName)},
		{Key: "color", Value: color},
	}
}

func NewMutableDirectedGraph(roots []string) *MutableDirectedGraph {
	return &MutableDirectedGraph{
		DirectedGraph: concrete.NewDirectedGraph(),
		nodesByName:   make(map[string]graph.Node),
		rootNodeNames: roots,
	}
}

type MutableDirectedGraph struct {
	*concrete.DirectedGraph

	nodesByName   map[string]graph.Node
	rootNodeNames []string
}

func (g *MutableDirectedGraph) AddNode(n *Node) error {
	if _, exists := g.nodesByName[n.UniqueName]; exists {
		return fmt.Errorf("node .UniqueName collision: %s", n.UniqueName)
	}

	g.nodesByName[n.UniqueName] = n
	g.DirectedGraph.AddNode(n)
	return nil
}

// RemoveNode deletes the given node from the graph,
// removing all of its outbound and inbound edges as well.
func (g *MutableDirectedGraph) RemoveNode(n *Node) {
	delete(g.nodesByName, n.UniqueName)
	g.DirectedGraph.RemoveNode(n)
}

func (g *MutableDirectedGraph) NodeByName(name string) (graph.Node, bool) {
	n, exists := g.nodesByName[name]
	return n, exists && g.DirectedGraph.Has(n)
}

// PruneOrphans recursively removes nodes with no inbound edges.
// Nodes marked as root nodes are ignored.
// Returns a list of recursively pruned nodes.
func (g *MutableDirectedGraph) PruneOrphans() []*Node {
	removed := []*Node{}

	for _, n := range g.nodesByName {
		node, ok := n.(*Node)
		if !ok {
			continue
		}
		if len(g.To(n)) > 0 {
			continue
		}
		if contains(node.UniqueName, g.rootNodeNames) {
			continue
		}

		g.RemoveNode(node)
		removed = append(removed, node)
	}

	if len(removed) == 0 {
		return []*Node{}
	}

	return append(removed, g.PruneOrphans()...)
}

func contains(needle string, haystack []string) bool {
	for _, str := range haystack {
		if needle == str {
			return true
		}
	}

	return false
}

// Copy creates a new graph instance, preserving all of the nodes and edges
// from the original graph. The nodes themselves are shallow copies from the
// original graph.
func (g *MutableDirectedGraph) Copy() *MutableDirectedGraph {
	newGraph := NewMutableDirectedGraph(g.rootNodeNames)

	// copy nodes
	for _, n := range g.Nodes() {
		newNode, ok := n.(*Node)
		if !ok {
			continue
		}

		if err := newGraph.AddNode(newNode); err != nil {
			// this should never happen: the only error that could occur is a node name collision,
			// which would imply that the original graph that we are copying had node-name collisions.
			panic(fmt.Errorf("unexpected error while copying graph: %v", err))
		}
	}

	// copy edges
	for _, n := range g.Nodes() {
		node, ok := n.(*Node)
		if !ok {
			continue
		}

		if _, exists := newGraph.NodeByName(node.UniqueName); !exists {
			continue
		}

		from := g.From(n)
		for _, to := range from {
			if newGraph.HasEdgeFromTo(n, to) {
				continue
			}

			newGraph.SetEdge(concrete.Edge{
				F: n,
				T: to,
			}, 0)
		}
	}

	return newGraph
}
