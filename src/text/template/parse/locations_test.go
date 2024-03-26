package parse

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_simple_tree(t *testing.T) {
	code := "{{ . }}"
	//       ^^^^^^^ List
	//          ^    Action -> Pipe -> Command -> Dot
	root := parseRoot(code)

	assert.Equal(t, Location{0, 7}, root.Location)
	actionNode := root.Nodes[0].(*ActionNode)
	assert.Equal(t, Location{3, 1}, actionNode.Location)
	cmd := actionNode.Pipe.Cmds[0]
	assert.Equal(t, Location{3, 1}, cmd.Location)
	dot := cmd.Args[0].(*DotNode)
	assert.Equal(t, Location{3, 1}, dot.Location)
}

func Test_simple_value(t *testing.T) {
	code := "{{ .Values.foobar }}"
	//       ^^^^^^^^^^^^^^^^^^^^ List (0; 20)
	//          ^^^^^^^           Action (3; 7)
	//          ^^^^^^^^^^^^^^^   Pipe (3; 15)
	//          ^^^^^^^           Command (3; 7)
	//                 ^^^^^^^    Field (10; 7)
	root := parseRoot(code)

	assert.Equal(t, Location{0, 20}, root.Location)
	actionNode := root.Nodes[0].(*ActionNode)
	assert.Equal(t, Location{3, 7}, actionNode.Location)
	pipeNode := actionNode.Pipe
	assert.Equal(t, Location{3, 15}, pipeNode.Location)
	cmd := pipeNode.Cmds[0]
	assert.Equal(t, Location{3, 7}, cmd.Location)
	fieldNode := cmd.Args[0].(*FieldNode)
	// TODO: Why doesn't location include "Values"?
	assert.Equal(t, Location{10, 7}, fieldNode.Location)
	assert.Equal(t, []string{"Values", "foobar"}, fieldNode.Ident)
}

func Test_only_comment(t *testing.T) {
	code := "{{/* comment */}}"
	//         ^^^^^^^^^^^^^^^ List (opening braces are not included)
	//         ^^^^^^^^^^^^^   Comment
	root := parseRoot(code)

	// TODO: Pos is 2 and not 0, because `lex.go:lexLeftDelim` ignores opening braces if followed by `/*`.
	//  These characters are not included in the length as well.
	assert.Equal(t, Location{2, 15}, root.Location)
	commentNode := root.Nodes[0].(*CommentNode)
	assert.Equal(t, Location{2, 13}, commentNode.Location)
}

func Test_Continue_Break(t *testing.T) {
	code := `{{ range .Values.annotations }}{{ continue }}{{ break }}{{ end }}`
	//       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ List (0; 65)
	//                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^    Range (9; 53)
	//                ^^^^^^^^^^^^^^^^^^^^                                     Pipe (9; 20)
	//                                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^    List (31; 31)
	//                                         ^^^^^^^^                        Continue (34; 8)
	//                                                       ^^^^^             Break (48; 5)
	root := parseRoot(code)

	assert.Equal(t, Location{0, 65}, root.Location)
	rangeNode := root.Nodes[0].(*RangeNode)
	// TODO: range doesn't include the `range` keyword
	// TODO: `nodeEnd` doesn't include closing braces, so length of `rangeNode` stops after `end` keyword
	assert.Equal(t, Location{9, 53}, rangeNode.Location)
	pipeNode := rangeNode.List.Nodes[0].(*PipeNode)
	assert.Equal(t, Location{9, 20}, pipeNode.Location)
	listNode := rangeNode.List.Nodes[1].(*ListNode)
	assert.Equal(t, Location{31, 31}, listNode.Location)
	continueNode := listNode.Nodes[0].(*CommandNode)
	assert.Equal(t, Location{34, 8}, continueNode.Location)
	breakNode := listNode.Nodes[1].(*CommandNode)
	assert.Equal(t, Location{48, 5}, breakNode.Location)
}

func parseRoot(code string) *ListNode {
	trees, _ := Parse("test", code, "{{", "}}")
	return trees["test"].Root
}
