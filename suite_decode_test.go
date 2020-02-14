package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

type (
	tDecode       func(string, interface{}) error
	decodeSubtest func(*testing.T, tDecode)
)

type decodeSuite struct {
	level      qry.DecodeLevel
	decodeOpts []qry.Option
}

func (ds decodeSuite) with(opts ...qry.Option) decodeSuite {
	res := decodeSuite{level: ds.level}
	if len(ds.decodeOpts) > 0 {
		res.decodeOpts = make([]qry.Option, len(ds.decodeOpts))
		copy(res.decodeOpts, ds.decodeOpts)
	}
	res.decodeOpts = append(res.decodeOpts, opts...)
	return res
}

func (ds decodeSuite) dumpTraceOnFailure(t *testing.T, trace *qry.TraceTree) {
	if t.Failed() {
		t.Logf("Decode Trace:\n%s\n", trace.Sdump())
	}
}

func (ds decodeSuite) runSubtest(t *testing.T, name string, subtest decodeSubtest) {
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
		assert.NotPanics(t, func() {
			subtest(t, func(input string, v interface{}) error {
				return decoder.Decode(ds.level, input, v, trace)
			})
		})
	})
}
