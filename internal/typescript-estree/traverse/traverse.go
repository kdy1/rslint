// Package traverse provides AST traversal utilities for TypeScript ESTree nodes.
package traverse

import (
	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// Traverse performs a depth-first traversal of the AST starting from the given node.
// The visitor's Enter method is called when entering each node, and Leave is called when leaving.
// Traversal can be controlled using VisitorAction return values.
func Traverse(node types.Node, visitor Visitor) {
	if node == nil || visitor == nil {
		return
	}
	traverse(node, nil, visitor)
}

// traverse is the internal recursive implementation of AST traversal.
func traverse(node types.Node, parent types.Node, visitor Visitor) VisitorAction {
	if node == nil {
		return Continue
	}

	// Call Enter callback
	action := visitor.Enter(node, parent)

	switch action {
	case Stop:
		return Stop
	case Skip:
		// Skip children but continue with siblings
		visitor.Leave(node, parent)
		return Continue
	case Continue:
		// Continue with children
	}

	// Traverse children
	children := GetChildNodes(node)
	for _, child := range children {
		if child != nil {
			if action := traverse(child, node, visitor); action == Stop {
				return Stop
			}
		}
	}

	// Call Leave callback
	visitor.Leave(node, parent)
	return Continue
}

// Walk is a convenience function that walks the AST with a simple function callback.
// It's equivalent to Traverse with a SimpleVisitor that only has an Enter callback.
func Walk(node types.Node, fn func(node types.Node, parent types.Node) VisitorAction) {
	Traverse(node, &SimpleVisitor{
		EnterFunc: fn,
	})
}

// WalkSimple is a convenience function that walks the AST with a simple function callback
// that doesn't need to control traversal flow. All nodes are visited.
func WalkSimple(node types.Node, fn func(node types.Node, parent types.Node)) {
	Traverse(node, &SimpleVisitor{
		EnterFunc: func(node types.Node, parent types.Node) VisitorAction {
			fn(node, parent)
			return Continue
		},
	})
}

// Find searches for the first node that matches the given predicate.
// Returns the matching node and its parent, or (nil, nil) if not found.
func Find(root types.Node, predicate func(node types.Node) bool) (types.Node, types.Node) {
	var found types.Node
	var foundParent types.Node

	Walk(root, func(node types.Node, parent types.Node) VisitorAction {
		if predicate(node) {
			found = node
			foundParent = parent
			return Stop
		}
		return Continue
	})

	return found, foundParent
}

// FindAll searches for all nodes that match the given predicate.
// Returns a slice of matching nodes.
func FindAll(root types.Node, predicate func(node types.Node) bool) []types.Node {
	var matches []types.Node

	WalkSimple(root, func(node types.Node, parent types.Node) {
		if predicate(node) {
			matches = append(matches, node)
		}
	})

	return matches
}

// FindByType searches for the first node of the given type.
// Returns the matching node and its parent, or (nil, nil) if not found.
func FindByType(root types.Node, nodeType string) (types.Node, types.Node) {
	return Find(root, func(node types.Node) bool {
		return node.Type() == nodeType
	})
}

// FindAllByType searches for all nodes of the given type.
// Returns a slice of matching nodes.
func FindAllByType(root types.Node, nodeType string) []types.Node {
	return FindAll(root, func(node types.Node) bool {
		return node.Type() == nodeType
	})
}

// Ancestors returns the chain of ancestor nodes from the given node to the root.
// The first element is the immediate parent, and the last element is the root.
// This requires traversing from the root to build the parent chain.
func Ancestors(root types.Node, target types.Node) []types.Node {
	if root == nil || target == nil {
		return nil
	}

	var ancestors []types.Node
	var currentPath []types.Node

	Walk(root, func(node types.Node, parent types.Node) VisitorAction {
		currentPath = append(currentPath, node)

		if node == target {
			// Found the target, copy the path (excluding the target itself)
			if len(currentPath) > 1 {
				ancestors = make([]types.Node, len(currentPath)-1)
				copy(ancestors, currentPath[:len(currentPath)-1])
				// Reverse to have immediate parent first
				for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
					ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
				}
			}
			return Stop
		}

		return Continue
	})

	// Clean up the path when leaving a node
	Traverse(root, &SimpleVisitor{
		LeaveFunc: func(node types.Node, parent types.Node) {
			if len(currentPath) > 0 && currentPath[len(currentPath)-1] == node {
				currentPath = currentPath[:len(currentPath)-1]
			}
		},
	})

	return ancestors
}

// HasAncestor checks if the target node has an ancestor matching the predicate.
// This requires traversing from the root to determine the parent chain.
func HasAncestor(root types.Node, target types.Node, predicate func(node types.Node) bool) bool {
	ancestors := Ancestors(root, target)
	for _, ancestor := range ancestors {
		if predicate(ancestor) {
			return true
		}
	}
	return false
}

// GetParent finds and returns the parent of the target node.
// Returns nil if the target is the root or not found.
func GetParent(root types.Node, target types.Node) types.Node {
	if root == nil || target == nil || root == target {
		return nil
	}

	var parent types.Node

	Walk(root, func(node types.Node, p types.Node) VisitorAction {
		if node == target {
			parent = p
			return Stop
		}
		return Continue
	})

	return parent
}

// DepthFirstIterator returns a channel that yields nodes in depth-first order.
// This is useful for processing nodes sequentially without recursion.
func DepthFirstIterator(root types.Node) <-chan types.Node {
	ch := make(chan types.Node)

	go func() {
		defer close(ch)
		WalkSimple(root, func(node types.Node, parent types.Node) {
			ch <- node
		})
	}()

	return ch
}

// Transform applies a transformation function to all nodes in the AST.
// The transformer function receives a node and should return a replacement node,
// or the same node if no transformation is needed.
// Note: This creates a new AST; the original is not modified.
func Transform(node types.Node, transformer func(node types.Node) types.Node) types.Node {
	if node == nil {
		return nil
	}
	return transformer(node)
}
