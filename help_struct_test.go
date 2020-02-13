package qry_test

import (
	"testing"

	"github.com/oligarch316/qry"
	"github.com/stretchr/testify/assert"
)

type (
	TestStructBasic struct {
		KeyA string
		KeyB string
	}

	TestStructStructBasic struct{ TestStructBasic }

	TestStructStarStructBasic struct{ *TestStructBasic }

	testStructUnexported struct {
		KeyA string
		KeyB string
	}

	testStructStructUnexported struct{ testStructUnexported }
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

func testStructs(t *testing.T) {
	// TODO: Make usable for both query and field levels (rather than just query as is current)
	base := newTest(
		configOptionsAs(qry.SetValueListVia(qry.SetAllowLiteral)),
		decodeLevelAs(qry.LevelQuery),
		inputAs("keyA=val%20A&keyB=val%20B&keyC=val%20C"),
	)

	t.Run("basic", func(t *testing.T) {
		var target TestStructBasic

		trace := base.require(t, &target)
		success := assert.Equal(t, "val A", target.KeyA)
		success = assert.Equal(t, "val B", target.KeyB) && success

		if !success {
			trace.log(t)
		}
	})

	t.Run("tag name", func(t *testing.T) {
		var target struct {
			KeyOne string `qry:"keyA"`
			KeyTwo string `qry:"keyB"`
		}

		trace := base.require(t, &target)
		success := assert.Equal(t, "val A", target.KeyOne)
		success = assert.Equal(t, "val B", target.KeyTwo) && success

		if !success {
			trace.log(t)
		}
	})

	t.Run("tag embed", func(t *testing.T) {
		t.Run("inner exported struct", func(t *testing.T) {
			t.Run("one level", func(t *testing.T) {
				var target struct {
					Embedded TestStructBasic `qry:",embed"`
				}

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.Embedded.KeyA)
				success = assert.Equal(t, "val B", target.Embedded.KeyB) && success

				if !success {
					trace.log(t)
				}
			})

			t.Run("two levels", func(t *testing.T) {
				var target struct {
					EmbeddedOne struct {
						EmbeddedTwo TestStructBasic `qry:",embed"`
					} `qry:",embed"`
				}

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.EmbeddedOne.EmbeddedTwo.KeyA)
				success = assert.Equal(t, "val B", target.EmbeddedOne.EmbeddedTwo.KeyB) && success

				if !success {
					trace.log(t)
				}
			})
		})

		t.Run("inner exported *struct", func(t *testing.T) {
			t.Run("one level", func(t *testing.T) {
				var target struct {
					Embedded *TestStructBasic `qry:",embed"`
				}

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.Embedded.KeyA)
				success = assert.Equal(t, "val B", target.Embedded.KeyB) && success

				if !success {
					trace.log(t)
				}
			})

			t.Run("two levels", func(t *testing.T) {
				var target struct {
					EmbeddedOne *struct {
						EmbeddedTwo *TestStructBasic `qry:",embed"`
					} `qry:",embed"`
				}

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.EmbeddedOne.EmbeddedTwo.KeyA)
				success = assert.Equal(t, "val B", target.EmbeddedOne.EmbeddedTwo.KeyB) && success

				if !success {
					trace.log(t)
				}
			})
		})
	})

	t.Run("anonymous embed", func(t *testing.T) {
		t.Run("inner unexported struct", func(t *testing.T) {
			t.Run("one level", func(t *testing.T) {
				var target struct{ testStructUnexported }

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.KeyA)
				success = assert.Equal(t, "val B", target.KeyB) && success

				if !success {
					trace.log(t)
				}
			})

			t.Run("two levels", func(t *testing.T) {
				var target struct{ testStructStructUnexported }

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.KeyA)
				success = assert.Equal(t, "val B", target.KeyB) && success

				if !success {
					trace.log(t)
				}
			})
		})

		t.Run("inner exported struct", func(t *testing.T) {
			t.Run("one level", func(t *testing.T) {
				var target struct{ TestStructBasic }

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.KeyA)
				success = assert.Equal(t, "val B", target.KeyB) && success

				if !success {
					trace.log(t)
				}
			})

			t.Run("two levels", func(t *testing.T) {
				var target struct{ TestStructStructBasic }

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.KeyA)
				success = assert.Equal(t, "val B", target.KeyB) && success

				if !success {
					trace.log(t)
				}
			})
		})

		t.Run("inner exported *struct", func(t *testing.T) {
			t.Run("one level", func(t *testing.T) {
				var target struct{ *TestStructBasic }

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.KeyA)
				success = assert.Equal(t, "val B", target.KeyB) && success

				if !success {
					trace.log(t)
				}
			})

			t.Run("two levels", func(t *testing.T) {
				var target struct{ *TestStructStarStructBasic }

				trace := base.require(t, &target)
				success := assert.Equal(t, "val A", target.KeyA)
				success = assert.Equal(t, "val B", target.KeyB) && success

				if !success {
					trace.log(t)
				}
			})
		})
	})
}
