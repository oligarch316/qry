package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
)

func testFauxLiteralErrors(t *testing.T, level qry.DecodeLevel) {
	t.Run("unescape", func(t *testing.T) {
		t.Skip("TODO: converter.Unescape error")
	})

	t.Run("array too small", func(t *testing.T) {
		t.Skip("TODO: insufficient target length error")
	})
}
