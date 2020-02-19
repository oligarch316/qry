package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
)

func testLiteralErrors(t *testing.T, level qry.DecodeLevel) {
	t.Run("convert", func(t *testing.T) {
		t.Skip("TODO: all possible strconv.XYZ errors")
	})

	t.Run("unescape", func(t *testing.T) {
		t.Skip("TODO: converter.Unescape error")
	})

	t.Run("unmarshal", func(t *testing.T) {
		t.Skip("TODO: unmarshaler error")
	})
}
