package qry

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

// Trace TODO
type Trace interface {
	Mark(DecodeLevel, string, reflect.Value)
	Child() Trace
}

func mergeTraces(traces []Trace) Trace {
	switch len(traces) {
	case 0:
		// No-op
		return MarkTrace(func(DecodeLevel, string, reflect.Value) {})
	case 1:
		return traces[0]
	}

	return TraceList(traces)
}

// TraceList TODO
type TraceList []Trace

// Mark TODO
func (tl TraceList) Mark(level DecodeLevel, input string, target reflect.Value) {
	for _, t := range tl {
		t.Mark(level, input, target)
	}
}

// Child TODO
func (tl TraceList) Child() Trace {
	res := make(TraceList, len(tl))
	for i, t := range tl {
		res[i] = t.Child()
	}
	return res
}

// MarkTrace TODO
type MarkTrace func(DecodeLevel, string, reflect.Value)

// Mark TODO
func (mt MarkTrace) Mark(level DecodeLevel, input string, target reflect.Value) {
	mt(level, input, target)
}

// Child TODO
func (mt MarkTrace) Child() Trace { return mt }

// TreeTrace TODO
type TreeTrace struct{ TreeTraceNode }

// NewTreeTrace TODO
func NewTreeTrace() *TreeTrace { return new(TreeTrace) }

func (tt *TreeTrace) String() string {
	if tt.Input == "" {
		return "trace tree"
	}

	return fmt.Sprintf("trace tree: %s", tt.TreeTraceNode.String())
}

// Dump TODO
func (tt TreeTrace) Dump() { tt.Fdump(os.Stdout) }

// Fdump TODO
func (tt TreeTrace) Fdump(w io.Writer) { dumpTree(w, LevelQuery, &tt.TreeTraceNode) }

// Sdump TODO
func (tt TreeTrace) Sdump() string {
	var buf bytes.Buffer
	tt.Fdump(&buf)
	return buf.String()
}

func dumpTree(w io.Writer, parentLevel DecodeLevel, node *TreeTraceNode) {
	adjLevel := node.Level
	if adjLevel >= LevelValueList {
		adjLevel--
	}

	var (
		pad    = strings.Repeat(" ", int(parentLevel-LevelQuery))
		offset = int(adjLevel - parentLevel)
	)

	for i := 0; i < offset; i++ {
		fmt.Fprintf(w, "%s\\\n", pad)
		pad = pad + " "
	}

	fmt.Fprintf(w, "%s%s\n", pad, node.String())
	for _, child := range node.Children {
		dumpTree(w, node.Level, child)
	}
}

// TreeTraceNode TODO
type TreeTraceNode struct {
	DecodeInfo
	Children []*TreeTraceNode
}

// Mark TODO
func (ttn *TreeTraceNode) Mark(level DecodeLevel, input string, target reflect.Value) {
	ttn.DecodeInfo = level.newInfo(input, target)
}

// Child TODO
func (ttn *TreeTraceNode) Child() Trace {
	res := new(TreeTraceNode)
	ttn.Children = append(ttn.Children, res)
	return res
}
