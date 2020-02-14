package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
)

type (
	tDecode       func(string, interface{}) error
	decodeSubtest func(*testing.T, tDecode)
)

type decodeSuite struct {
	level      qry.DecodeLevel
	decodeOpts []qry.Option
}

func (ds decodeSuite) dumpTraceOnFailure(t *testing.T, trace *qry.TraceTree) {
	if t.Failed() {
		t.Logf("Decode Trace:\n%s\n", trace.Sdump())
	}
}

func (ds decodeSuite) run(t *testing.T, name string, subtest decodeSubtest) {
	t.Run(name, func(t *testing.T) {
		// Setup
		var (
			decoder = qry.NewDecoder(ds.decodeOpts...)
			trace   = qry.NewTraceTree()
		)

		// Teardown
		if testing.Verbose() {
			// Use defer here to ensure trace dump in FailNow() situations
			defer ds.dumpTraceOnFailure(t, trace)
		}

		// Test
		subtest(t, func(input string, v interface{}) error {
			return decoder.Decode(ds.level, input, v, trace)
		})
	})
}
