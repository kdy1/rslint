package traverse

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/typescript-estree/types"
)

// Helper function to create a simple AST for testing
func createTestAST() *types.Program {
	xIdentifier := &types.Identifier{
		BaseNode: types.BaseNode{NodeType: "Identifier"},
		Name:     "x",
	}
	var literal42 types.Expression = &types.SimpleLiteral{
		BaseNode: types.BaseNode{NodeType: "Literal"},
		Value:    float64(42),
		Raw:      "42",
	}
	xIdentifier2 := &types.Identifier{
		BaseNode: types.BaseNode{NodeType: "Identifier"},
		Name:     "x",
	}
	var literal10 types.Expression = &types.SimpleLiteral{
		BaseNode: types.BaseNode{NodeType: "Literal"},
		Value:    float64(10),
		Raw:      "10",
	}

	return &types.Program{
		BaseNode: types.BaseNode{NodeType: "Program"},
		Body: []types.Statement{
			&types.VariableDeclaration{
				BaseNode: types.BaseNode{NodeType: "VariableDeclaration"},
				Kind:     "const",
				Declarations: []types.VariableDeclarator{
					{
						BaseNode: types.BaseNode{NodeType: "VariableDeclarator"},
						ID:       xIdentifier,
						Init:     &literal42,
					},
				},
			},
			&types.ExpressionStatement{
				BaseNode: types.BaseNode{NodeType: "ExpressionStatement"},
				Expression: &types.BinaryExpression{
					BaseNode: types.BaseNode{NodeType: "BinaryExpression"},
					Operator: "+",
					Left:     xIdentifier2,
					Right:    literal10,
				},
			},
		},
	}
}

func TestTraverse(t *testing.T) {
	ast := createTestAST()

	visited := []string{}
	visitor := &SimpleVisitor{
		EnterFunc: func(node types.Node, parent types.Node) VisitorAction {
			visited = append(visited, node.Type())
			return Continue
		},
	}

	Traverse(ast, visitor)

	// Check that all nodes were visited
	expected := []string{
		"Program",
		"VariableDeclaration",
		"VariableDeclarator",
		"Identifier",
		"Literal",
		"ExpressionStatement",
		"BinaryExpression",
		"Identifier",
		"Literal",
	}

	if len(visited) != len(expected) {
		t.Errorf("Expected %d nodes visited, got %d", len(expected), len(visited))
	}

	for i, nodeType := range expected {
		if i >= len(visited) {
			break
		}
		if visited[i] != nodeType {
			t.Errorf("Node %d: expected %s, got %s", i, nodeType, visited[i])
		}
	}
}

func TestTraverseWithStop(t *testing.T) {
	ast := createTestAST()

	visited := []string{}
	visitor := &SimpleVisitor{
		EnterFunc: func(node types.Node, parent types.Node) VisitorAction {
			visited = append(visited, node.Type())
			// Stop after visiting the first VariableDeclaration
			if node.Type() == "VariableDeclaration" {
				return Stop
			}
			return Continue
		},
	}

	Traverse(ast, visitor)

	// Should only have visited Program and VariableDeclaration
	if len(visited) != 2 {
		t.Errorf("Expected 2 nodes visited, got %d: %v", len(visited), visited)
	}

	if visited[0] != "Program" || visited[1] != "VariableDeclaration" {
		t.Errorf("Unexpected nodes visited: %v", visited)
	}
}

func TestTraverseWithSkip(t *testing.T) {
	ast := createTestAST()

	visited := []string{}
	visitor := &SimpleVisitor{
		EnterFunc: func(node types.Node, parent types.Node) VisitorAction {
			visited = append(visited, node.Type())
			// Skip children of VariableDeclaration
			if node.Type() == "VariableDeclaration" {
				return Skip
			}
			return Continue
		},
	}

	Traverse(ast, visitor)

	// Should not visit children of VariableDeclaration
	containsVariableDeclarator := false
	for _, nodeType := range visited {
		if nodeType == "VariableDeclarator" {
			containsVariableDeclarator = true
			break
		}
	}

	if containsVariableDeclarator {
		t.Error("Should not have visited VariableDeclarator after Skip action")
	}

	// Should still visit ExpressionStatement
	containsExpressionStatement := false
	for _, nodeType := range visited {
		if nodeType == "ExpressionStatement" {
			containsExpressionStatement = true
			break
		}
	}

	if !containsExpressionStatement {
		t.Error("Should have visited ExpressionStatement after Skip action")
	}
}

func TestWalk(t *testing.T) {
	ast := createTestAST()

	count := 0
	Walk(ast, func(node types.Node, parent types.Node) VisitorAction {
		count++
		return Continue
	})

	if count != 9 {
		t.Errorf("Expected 9 nodes, got %d", count)
	}
}

func TestWalkSimple(t *testing.T) {
	ast := createTestAST()

	count := 0
	WalkSimple(ast, func(node types.Node, parent types.Node) {
		count++
	})

	if count != 9 {
		t.Errorf("Expected 9 nodes, got %d", count)
	}
}

func TestFind(t *testing.T) {
	ast := createTestAST()

	// Find BinaryExpression
	node, parent := Find(ast, func(node types.Node) bool {
		return node.Type() == "BinaryExpression"
	})

	if node == nil {
		t.Error("Expected to find BinaryExpression")
	}

	if node.Type() != "BinaryExpression" {
		t.Errorf("Expected BinaryExpression, got %s", node.Type())
	}

	if parent == nil || parent.Type() != "ExpressionStatement" {
		t.Error("Expected parent to be ExpressionStatement")
	}
}

func TestFindNotFound(t *testing.T) {
	ast := createTestAST()

	// Try to find a node that doesn't exist
	node, parent := Find(ast, func(node types.Node) bool {
		return node.Type() == "ForStatement"
	})

	if node != nil {
		t.Error("Expected not to find ForStatement")
	}

	if parent != nil {
		t.Error("Expected parent to be nil")
	}
}

func TestFindAll(t *testing.T) {
	ast := createTestAST()

	// Find all Identifier nodes
	nodes := FindAll(ast, func(node types.Node) bool {
		return node.Type() == "Identifier"
	})

	if len(nodes) != 2 {
		t.Errorf("Expected 2 Identifier nodes, got %d", len(nodes))
	}

	for _, node := range nodes {
		if node.Type() != "Identifier" {
			t.Errorf("Expected Identifier, got %s", node.Type())
		}
	}
}

func TestFindByType(t *testing.T) {
	ast := createTestAST()

	node, parent := FindByType(ast, "VariableDeclaration")

	if node == nil {
		t.Error("Expected to find VariableDeclaration")
	}

	if node.Type() != "VariableDeclaration" {
		t.Errorf("Expected VariableDeclaration, got %s", node.Type())
	}

	if parent == nil || parent.Type() != "Program" {
		t.Error("Expected parent to be Program")
	}
}

func TestFindAllByType(t *testing.T) {
	ast := createTestAST()

	nodes := FindAllByType(ast, "Literal")

	if len(nodes) != 2 {
		t.Errorf("Expected 2 Literal nodes, got %d", len(nodes))
	}
}

func TestGetParent(t *testing.T) {
	ast := createTestAST()

	// Find a BinaryExpression
	binExpr, _ := FindByType(ast, "BinaryExpression")
	if binExpr == nil {
		t.Fatal("Could not find BinaryExpression")
	}

	// Get its parent
	parent := GetParent(ast, binExpr)
	if parent == nil {
		t.Error("Expected parent to be non-nil")
	}

	if parent.Type() != "ExpressionStatement" {
		t.Errorf("Expected parent to be ExpressionStatement, got %s", parent.Type())
	}
}

func TestGetParentOfRoot(t *testing.T) {
	ast := createTestAST()

	parent := GetParent(ast, ast)
	if parent != nil {
		t.Error("Expected parent of root to be nil")
	}
}

func TestTypedVisitor(t *testing.T) {
	ast := createTestAST()

	identifierCount := 0
	literalCount := 0

	visitor := &TypedVisitor{
		Visitors: map[string]func(node types.Node, parent types.Node) VisitorAction{
			"Identifier": func(node types.Node, parent types.Node) VisitorAction {
				identifierCount++
				return Continue
			},
			"Literal": func(node types.Node, parent types.Node) VisitorAction {
				literalCount++
				return Continue
			},
		},
	}

	Traverse(ast, visitor)

	if identifierCount != 2 {
		t.Errorf("Expected 2 identifiers, got %d", identifierCount)
	}

	if literalCount != 2 {
		t.Errorf("Expected 2 literals, got %d", literalCount)
	}
}

func TestVisitorLeave(t *testing.T) {
	ast := createTestAST()

	entered := []string{}
	left := []string{}

	visitor := &SimpleVisitor{
		EnterFunc: func(node types.Node, parent types.Node) VisitorAction {
			entered = append(entered, node.Type())
			return Continue
		},
		LeaveFunc: func(node types.Node, parent types.Node) {
			left = append(left, node.Type())
		},
	}

	Traverse(ast, visitor)

	if len(entered) != len(left) {
		t.Errorf("Expected same number of enter and leave calls, got enter=%d, leave=%d", len(entered), len(left))
	}

	// Leave should be called in reverse depth-first order (children before parents)
	// This means the last node entered should be the first node left
	// But this is only true for leaf nodes, not for the entire tree
	// A simple check is to verify that all nodes that were entered were also left
	enteredMap := make(map[string]int)
	leftMap := make(map[string]int)

	for _, nodeType := range entered {
		enteredMap[nodeType]++
	}
	for _, nodeType := range left {
		leftMap[nodeType]++
	}

	for nodeType, count := range enteredMap {
		if leftMap[nodeType] != count {
			t.Errorf("Node type %s: entered %d times but left %d times", nodeType, count, leftMap[nodeType])
		}
	}
}

func TestDepthFirstIterator(t *testing.T) {
	ast := createTestAST()

	count := 0
	for node := range DepthFirstIterator(ast) {
		if node == nil {
			t.Error("Received nil node from iterator")
		}
		count++
	}

	if count != 9 {
		t.Errorf("Expected 9 nodes from iterator, got %d", count)
	}
}

func TestTraverseNil(t *testing.T) {
	// Should not panic with nil node
	Traverse(nil, &SimpleVisitor{})
}

func TestTraverseNilVisitor(t *testing.T) {
	ast := createTestAST()
	// Should not panic with nil visitor
	Traverse(ast, nil)
}
