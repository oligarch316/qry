package qry

import (
	"reflect"

	"github.com/disiqueira/gotree"
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

func (tt TraceTree) String() string {
	root := gotree.New("Decode Trace")
	tt.addToStringDump(root)
	return root.Print()
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

func (ttn TraceTreeNode) addToStringDump(dump gotree.Tree) {
	me := dump.Add(ttn.String())
	for _, child := range ttn.Children {
		child.addToStringDump(me)
	}
}
