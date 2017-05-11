// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package vm

import (
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

var testCodeGenPrograms = []struct {
	name   string
	source string
	prog   []instr // expected bytecode
}{
	// Composite literals require too many explicit conversions.
	{"simple line counter",
		"counter line_count\n/$/ { line_count++\n }\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 7},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"count a",
		"counter a_count\n/a$/ { a_count++\n }\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 7},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"strptime and capref",
		"counter foo\n" +
			"/(.*)/ { strptime($1, \"2006-01-02T15:04:05\")\n" +
			"foo++\n}\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 11},
			instr{setmatched, false},
			instr{push, 0},
			instr{capref, 0},
			instr{str, 0},
			instr{strptime, 2},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"strptime and named capref",
		"counter foo\n" +
			"/(?P<date>.*)/ { strptime($date, \"2006-01-02T15:04:05\")\n" +
			"foo++\n }\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 11},
			instr{setmatched, false},
			instr{push, 0},
			instr{capref, 1},
			instr{str, 0},
			instr{strptime, 2},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"inc by and set",
		"counter foo\ncounter bar\n" +
			"/(.*)/ {\n" +
			"foo += $1\n" +
			"bar = $1\n" +
			"}\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 17},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{mload, 0},
			instr{dload, 0},
			instr{push, 0},
			instr{capref, 0},
			instr{iadd, nil},
			instr{iset, nil},
			instr{mload, 1},
			instr{dload, 0},
			instr{push, 0},
			instr{capref, 0},
			instr{iset, nil},
			instr{setmatched, true}}},
	{"cond expr gt",
		"counter foo\n" +
			"1 > 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, 1},
			instr{jnm, 9},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"cond expr lt",
		"counter foo\n" +
			"1 < 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, -1},
			instr{jnm, 9},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"cond expr eq",
		"counter foo\n" +
			"1 == 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, 0},
			instr{jnm, 9},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"cond expr le",
		"counter foo\n" +
			"1 <= 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, 1},
			instr{jm, 9},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"cond expr ge",
		"counter foo\n" +
			"1 >= 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, -1},
			instr{jm, 9},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"cond expr ne",
		"counter foo\n" +
			"1 != 0 {\n" +
			"  foo++\n" +
			"}\n",
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, 0},
			instr{jm, 9},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"nested cond",
		"counter foo\n" +
			"/(.*)/ {\n" +
			"  $1 <= 1 {\n" +
			"    foo++\n" +
			"  }\n" +
			"}\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 14},
			instr{setmatched, false},
			instr{push, 0},
			instr{capref, 0},
			instr{push, 1},
			instr{cmp, 1},
			instr{jm, 13},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true},
			instr{setmatched, true}}},
	{"deco",
		"counter foo\n" +
			"counter bar\n" +
			"def fooWrap {\n" +
			"  /.*/ {\n" +
			"    foo++\n" +
			"    next\n" +
			"  }\n" +
			"}\n" +
			"" +
			"@fooWrap { bar++\n }\n",
		[]instr{
			instr{match, 0},
			instr{jnm, 10},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{mload, 1},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"length",
		"len(\"foo\") > 0 {\n" +
			"}\n",
		[]instr{
			instr{str, 0},
			instr{length, 1},
			instr{push, 0},
			instr{cmp, 1},
			instr{jnm, 7},
			instr{setmatched, false},
			instr{setmatched, true}}},
	{"bitwise", `
1 & 7 ^ 15 | 8
~ 16 << 2
1 >> 20
`,
		[]instr{
			instr{push, 1},
			instr{push, 7},
			instr{and, nil},
			instr{push, 15},
			instr{xor, nil},
			instr{push, 8},
			instr{or, nil},
			instr{push, 16},
			instr{not, nil},
			instr{push, 2},
			instr{shl, nil},
			instr{push, 1},
			instr{push, 20},
			instr{shr, nil}}},
	{"pow", `
counter a
counter b
a ** b
`,
		[]instr{
			instr{mload, 0},
			instr{dload, 0},
			instr{mload, 1},
			instr{dload, 0},
			instr{ipow, nil}}},
	{"indexed expr", `
counter a by b
a["string"]++
`,
		[]instr{
			instr{str, 0},
			instr{mload, 0},
			instr{dload, 1},
			instr{inc, nil}}},
	{"strtol", `
strtol("deadbeef", 16)
`,
		[]instr{
			instr{str, 0},
			instr{push, 16},
			instr{strtol, 2}}},
	{"float", `
20.0
`,
		[]instr{
			instr{push, 20.0}}},
	{"otherwise", `
counter a
otherwise {
	a++
}
`,
		[]instr{
			instr{otherwise, nil},
			instr{jnm, 7},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true}}},
	{"cond else",
		`counter foo
counter bar
1 > 0 {
  foo++
} else {
  bar++
}`,
		[]instr{
			instr{push, 1},
			instr{push, 0},
			instr{cmp, 1},
			instr{jnm, 10},
			instr{setmatched, false},
			instr{mload, 0},
			instr{dload, 0},
			instr{inc, nil},
			instr{setmatched, true},
			instr{jmp, 13},
			instr{mload, 1},
			instr{dload, 0},
			instr{inc, nil},
		},
	},
	{"mod",
		`
3 % 1
`,
		[]instr{
			instr{push, 3},
			instr{push, 1},
			instr{imod, nil},
		},
	},
	{"del", `
counter a by b
del a["string"]
`,
		[]instr{
			instr{str, 0},
			instr{mload, 0},
			instr{del, 1}},
	},
}

func TestCodegen(t *testing.T) {
	for _, tc := range testCodeGenPrograms {
		t.Logf("Starting %s", tc.name)
		ast, err := Parse(tc.name, strings.NewReader(tc.source))
		if err != nil {
			t.Fatalf("Unexpected parse failure in %q: %s", tc.name, err)
		}
		err = Check(ast)
		if err != nil {
			t.Fatalf("Unexpected check failure in %q: %s", tc.name, err)
		}
		obj, err := CodeGen(tc.name, ast)
		if err != nil {
			t.Errorf("Compile errors for %q:\n%q", tc.name, err)
			continue
		}
		diff := pretty.Compare(tc.prog, obj.prog)
		if len(diff) > 0 {
			t.Errorf("%s: VM prog doesn't match.\n%s", tc.name, diff)
		}
	}
}
