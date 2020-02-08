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
		return TraceMarker(func(DecodeLevel, string, reflect.Value) {})
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

// TraceMarker TODO
type TraceMarker func(DecodeLevel, string, reflect.Value)

// Mark TODO
func (tm TraceMarker) Mark(level DecodeLevel, input string, target reflect.Value) {
	tm(level, input, target)
}

// Child TODO
func (tm TraceMarker) Child() Trace { return tm }

// TraceTree TODO
type TraceTree struct{ TraceTreeNode }

// NewTraceTree TODO
func NewTraceTree() *TraceTree { return new(TraceTree) }

func (tt *TraceTree) String() string {
	if tt.Input == "" {
		return "trace tree"
	}

	return fmt.Sprintf("trace tree: %s", tt.TraceTreeNode.String())
}

// Dump TODO
func (tt TraceTree) Dump() { tt.Fdump(os.Stdout) }

// Fdump TODO
func (tt TraceTree) Fdump(w io.Writer) { dumpTree(w, LevelQuery, &tt.TraceTreeNode) }

// Sdump TODO
func (tt TraceTree) Sdump() string {
	var buf bytes.Buffer
	tt.Fdump(&buf)
	return buf.String()
}

// TODO: this is wrong!
func dumpTree(w io.Writer, parentLevel DecodeLevel, node *TraceTreeNode) {
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

// TraceTreeNode TODO
type TraceTreeNode struct {
	DecodeInfo
	Children []*TraceTreeNode
}

// Mark TODO
func (ttn *TraceTreeNode) Mark(level DecodeLevel, input string, target reflect.Value) {
	ttn.DecodeInfo = level.newInfo(input, target)
}

// Child TODO
func (ttn *TraceTreeNode) Child() Trace {
	res := new(TraceTreeNode)
	ttn.Children = append(ttn.Children, res)
	return res
}
