package utils

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestUtils(t *testing.T) {
	g := Goblin(t)

	g.Describe("ARange", func() {
		g.It("Should produce a range", func() {
			rg := ARange(10, 20, 10)
			g.Assert(len(rg)).Equal(2)
			g.Assert(rg[0]).Equal(10)
			g.Assert(rg[1]).Equal(20)
		})
	})
}
