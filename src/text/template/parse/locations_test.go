package parse

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_simple_tree(t *testing.T) {
	code := "{{ . }}"
	//       ^^^^^^^ List (0; 7)
	//          ^    └- Action (3; 1)
	//          ^       └- Pipe (3; 1)
	//          ^          └- Command -> Dot (3; 1)
	root := parseRoot(t, code)

	assert.Equal(t, Location{0, 7}, root.Location)
	actionNode := root.Nodes[0].(*ActionNode)
	assert.Equal(t, Location{3, 1}, actionNode.Location)
	pipeNode := actionNode.Pipe
	assert.Equal(t, Location{3, 1}, pipeNode.Location)
	cmd := actionNode.Pipe.Cmds[0]
	assert.Equal(t, Location{3, 1}, cmd.Location)
	dot := cmd.Args[0].(*DotNode)
	assert.Equal(t, Location{3, 1}, dot.Location)
}

func Test_simple_value(t *testing.T) {
	code := "{{ .Values.foobar }}"
	//       ^^^^^^^^^^^^^^^^^^^^ List (0; 20)
	//          ^^^^^^^           └- Action (3; 7)
	//          ^^^^^^^^^^^^^^       └- Pipe (3; 14)
	//          ^^^^^^^^^^^^^^          └- Command (3; 14)
	//          ^^^^^^^^^^^^^^             └- Field (3; 14)
	root := parseRoot(t, code)

	assert.Equal(t, Location{0, 20}, root.Location)
	actionNode := root.Nodes[0].(*ActionNode)
	assert.Equal(t, Location{3, 7}, actionNode.Location)
	pipeNode := actionNode.Pipe
	assert.Equal(t, Location{3, 14}, pipeNode.Location)
	cmd := pipeNode.Cmds[0]
	assert.Equal(t, Location{3, 14}, cmd.Location)
	fieldNode := cmd.Args[0].(*FieldNode)
	assert.Equal(t, Location{3, 14}, fieldNode.Location)
	assert.Equal(t, []string{"Values", "foobar"}, fieldNode.Ident)
}

func Test_only_comment(t *testing.T) {
	code := "{{/* comment */}}"
	//         ^^^^^^^^^^^^^^^ List (opening braces are not included)
	//         ^^^^^^^^^^^^^   └- Comment
	root := parseRoot(t, code)

	// TODO: Pos is 2 and not 0, because `lex.go:lexLeftDelim` ignores opening braces if followed by `/*`.
	//  These characters are not included in the length as well.
	assert.Equal(t, Location{2, 15}, root.Location)
	commentNode := root.Nodes[0].(*CommentNode)
	assert.Equal(t, Location{2, 13}, commentNode.Location)
}

func Test_Continue_Break(t *testing.T) {
	code := `{{ range .Values.annotations }}{{ continue }}{{ break }}{{ end }}`
	//       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ List (0; 65)
	//                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^    └- Range (9; 53)
	//                ^^^^^^^^^^^^^^^^^^^                                         |- Pipe (9; 19)
	//                                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^       └- List (31; 31)
	//                                         ^^^^^^^^                              |- Continue (34; 8)
	//                                                       ^^^^^                   └- Break (48; 5)
	root := parseRoot(t, code)

	assert.Equal(t, Location{0, 65}, root.Location)
	rangeNode := root.Nodes[0].(*RangeNode)
	// TODO: range doesn't include the `range` keyword
	// TODO: `nodeEnd` doesn't include closing braces, so length of `rangeNode` stops after `end` keyword
	assert.Equal(t, Location{9, 53}, rangeNode.Location)
	pipeNode := rangeNode.Pipe
	assert.Equal(t, Location{9, 19}, pipeNode.Location)
	listNode := rangeNode.List
	assert.Equal(t, Location{31, 31}, listNode.Location)
	continueNode := listNode.Nodes[0].(*ContinueNode)
	assert.Equal(t, Location{34, 8}, continueNode.Location)
	breakNode := listNode.Nodes[1].(*BreakNode)
	assert.Equal(t, Location{48, 5}, breakNode.Location)
}

func Test_range_with_variables(t *testing.T) {
	code := `{{ range $key, $value := .Values.annotations }}{{ end }}`
	//       ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ List (0; 56)
	//                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^    |- Range (9; 44)
	//                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^             |  └- Pipe (9; 35)
	//                ^^^^                                            |     |- Variable (9; 4)
	//                      ^^^^^^                                    |     |- Variable (15; 6)
	//                                ^^^^^^^^^^^^^^^^^^^             |     └- Command (25; 19)
	//                                                     ^^^^^^^    └- List (47; 6)
	root := parseRoot(t, code)

	assert.Equal(t, Location{0, 56}, root.Location)
	rangeNode := root.Nodes[0].(*RangeNode)
	assert.Equal(t, Location{9, 44}, rangeNode.Location)
	pipeNode := rangeNode.Pipe
	assert.Equal(t, Location{9, 35}, pipeNode.Location)
	assert.Len(t, rangeNode.Pipe.Decl, 2)
	var1 := rangeNode.Pipe.Decl[0]
	assert.Equal(t, Location{9, 4}, var1.Location)
	var2 := rangeNode.Pipe.Decl[1]
	assert.Equal(t, Location{15, 6}, var2.Location)
	assert.Len(t, rangeNode.Pipe.Cmds, 1)
	cmdNode := rangeNode.Pipe.Cmds[0]
	assert.Equal(t, Location{25, 19}, cmdNode.Location)
	listNode := rangeNode.List
	assert.Equal(t, Location{47, 6}, listNode.Location)
}

func Test_function(t *testing.T) {
	code := `{{ foo "bar" }}`
	//       ^^^^^^^^^^^^^^^ List (0; 15)
	//          ^^^          └- Action (3; 3)
	//          ^^^^^^^^^       └- Pipe (3; 9)
	//          ^^^^^^^^^          └- Command (3; 9)
	//          ^^^      		      |- Identifier (3; 3)
	//              ^^^^^		      └- String (7; 5)

	trees, err := Parse("test", code, "{{", "}}", map[string]any{"foo": strings.ToUpper})
	assert.NoError(t, err)
	root := trees["test"].Root

	assert.Equal(t, Location{0, 15}, root.Location)
	actionNode := root.Nodes[0].(*ActionNode)
	assert.Equal(t, Location{3, 3}, actionNode.Location)
	pipeNode := actionNode.Pipe
	assert.Equal(t, Location{3, 9}, pipeNode.Location)
	cmdNode := pipeNode.Cmds[0]
	assert.Equal(t, Location{3, 9}, cmdNode.Location)
	identNode := cmdNode.Args[0].(*IdentifierNode)
	assert.Equal(t, Location{3, 3}, identNode.Location)
	stringNode := cmdNode.Args[1].(*StringNode)
	assert.Equal(t, Location{7, 5}, stringNode.Location)
}

func parseRoot(t *testing.T, code string) *ListNode {
	trees, err := Parse("test", code, "{{", "}}")

	assert.NoError(t, err)
	return trees["test"].Root
}
