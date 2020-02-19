package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
)

type (
	TestStructBasic struct {
		KeyA string
		KeyB string
	}
)

func testStructTagErrors(t *testing.T, level qry.DecodeLevel) {
	t.Run("tag embed", func(t *testing.T) {
		base := newTest(
			decodeLevelAs(level),
			inputAs("xyz"),
			checkDecodeError(
				assertDecodeLevel(level),
			),
			checkStructFieldError(),
		)

		t.Run("non-empty name", func(t *testing.T) {
			var target struct {
				Item string `qry:"item,embed"`
			}
			base.with(errorMessageAs("mutually exclusive non-empty name and embed directives")).require(t, &target)
		})

		t.Run("anonymous field", func(t *testing.T) {
			var target struct {
				TestStructBasic `qry:",embed"`
			}
			base.with(errorMessageAs("unecessary embed directive on anonymous field")).require(t, &target)
		})

		t.Run("unexported field", func(t *testing.T) {
			var target struct {
				item string `qry:",embed"`
			}
			base.with(errorMessageAs("embed directive on unexported field")).require(t, &target)
		})

		t.Run("neither pointer nor struct", func(t *testing.T) {
			var target struct {
				Item string `qry:",embed"`
			}
			base.with(errorMessageAs("embed directive on invalid type")).require(t, &target)
		})

		t.Run("pointer to non-struct", func(t *testing.T) {
			var target struct {
				Item *string `qry:",embed"`
			}
			base.with(errorMessageAs("embed directive on pointer to invalid type")).require(t, &target)
		})
	})
}
