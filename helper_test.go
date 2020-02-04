package qry_test

import (
	"errors"
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

type testUnmarshalerData struct {
	forcedErr error
	called    bool
	val       string
}

func (tud *testUnmarshalerData) doUnmarshal(text []byte) error {
	tud.called = true
	tud.val = string(text)
	return tud.forcedErr
}

func (tud testUnmarshalerData) assert(t *testing.T, expected string) bool {
	if !assert.True(t, tud.called) {
		return false
	}
	return assert.Equal(t, expected, tud.val)
}

type (
	TestStringAsserter interface{ assert(*testing.T, string) bool }

	TestCustomString   string
	TestUnmarshaler    struct{ testUnmarshalerData }
	TestRawUnmarshaler struct{ testUnmarshalerData }
)

func (tcs TestCustomString) assert(t *testing.T, expected string) bool {
	return assert.Equal(t, expected, string(tcs))
}
func (tu *TestUnmarshaler) UnmarshalText(text []byte) error        { return tu.doUnmarshal(text) }
func (tru *TestRawUnmarshaler) UnmarshalRawText(text []byte) error { return tru.doUnmarshal(text) }

type testTrace struct{ *qry.TreeTrace }

func (tt testTrace) log(t *testing.T) {
	if testing.Verbose() {
		t.Logf("Decode Trace:\n%s\n", tt.Sdump())
	}
}

type (
	decodeTest struct {
		configOpts       []qry.Option
		decodeLevel      qry.DecodeLevel
		input            string
		expectedErr      func(*testing.T, error) bool
		expectedErrLevel qry.DecodeLevel
	}

	testOpt func(*decodeTest)
)

func newTest(opts ...testOpt) (res decodeTest) {
	for _, opt := range opts {
		opt(&res)
	}
	return
}

func (dt decodeTest) with(opts ...testOpt) decodeTest {
	res := decodeTest{
		input:            dt.input,
		decodeLevel:      dt.decodeLevel,
		expectedErr:      dt.expectedErr,
		expectedErrLevel: dt.expectedErrLevel,
	}

	if dt.configOpts != nil {
		res.configOpts = make([]qry.Option, len(dt.configOpts))
		copy(res.configOpts, dt.configOpts)
	}

	for _, opt := range opts {
		opt(&res)
	}

	return res
}

func (dt decodeTest) assert(t *testing.T, out interface{}) (testTrace, bool) {
	var (
		decoder = qry.NewDecoder(dt.configOpts...)
		trace   = testTrace{qry.NewTreeTrace()}
	)

	actualErr := decoder.Decode(dt.decodeLevel, dt.input, out, trace)

	if dt.expectedErr == nil {
		return trace, assert.NoError(t, actualErr)
	}

	var decodeErr qry.DecodeError
	if !errors.As(actualErr, &decodeErr) {
		return trace, assert.Failf(t, "error is not a DecodeError", "%s (%T)", actualErr, actualErr)
	}

	levelCheck := assert.Equal(t, dt.expectedErrLevel, decodeErr.Level)

	return trace, dt.expectedErr(t, decodeErr.Unwrap()) && levelCheck
}

func (dt decodeTest) require(t *testing.T, out interface{}) testTrace {
	trace, success := dt.assert(t, out)
	if !success {
		trace.log(t)
		t.FailNow()
	}
	return trace
}

func configOptionsAs(opts ...qry.Option) testOpt {
	return func(d *decodeTest) { d.configOpts = opts }
}

func configOptionsAppend(opts ...qry.Option) testOpt {
	return func(d *decodeTest) { d.configOpts = append(d.configOpts, opts...) }
}

func inputAs(input string) testOpt {
	return func(d *decodeTest) { d.input = input }
}

func decodeLevelAs(level qry.DecodeLevel) testOpt {
	return func(d *decodeTest) { d.decodeLevel = level }
}

func errorLevelAs(level qry.DecodeLevel) testOpt {
	return func(d *decodeTest) { d.expectedErrLevel = level }
}

func errorAs(errString string, msgAndArgs ...interface{}) testOpt {
	checkErr := func(t *testing.T, err error) bool { return assert.EqualError(t, err, errString, msgAndArgs...) }
	return func(d *decodeTest) { d.expectedErr = checkErr }
}

func errorLike(errRx string, msgAndArgs ...interface{}) testOpt {
	checkErr := func(t *testing.T, err error) bool { return assert.Regexp(t, errRx, err.Error(), msgAndArgs...) }
	return func(d *decodeTest) { d.expectedErr = checkErr }
}
