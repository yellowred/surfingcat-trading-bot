package trading

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestSupertest(t *testing.T) {
	g := Goblin(t)

	g.Describe("Trading Strategy DIP", func() {

		g.It("Should be no action on a flat trend", func() {
			config := map[string]string{"param1": "1", "param2": "apple"}
			variableValues := map[string][]string{"param1": []string{"1", "2", "3"}}
			r := TestConfigFactory(config, variableValues)
			g.Assert(len(r)).Equal(3)
			g.Assert(r[2]["param1"]).Equal("3")
		})

	})
}
