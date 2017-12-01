package utils

import (
	. "github.com/franela/goblin"
	"testing"
)

func TestUtils(t *testing.T) {
	g := Goblin(t)

	g.Describe("ARange", func() {

		g.It("Should produce a range", func() {

			rg := ARange(10, 20, 10)
			g.Assert(len(rg)).Equal(2)
			g.Assert(rg[0] == 10).IsTrue()
			g.Assert(rg[1] == 20).IsTrue()
		})
	})
}
