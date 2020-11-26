package options

import (
	"testing"
)

func TestIPNetSliceContains(t *testing.T) {
	cases := []struct {
		inputNets []string
		inputIP   string
		want      bool
	}{
		{
			[]string{"192.168.0.1/24"},
			"192.168.0.22",
			true,
		},
		{
			[]string{"192.168.0.1/24"},
			"192.168.1.22",
			false,
		},
		{
			[]string{"192.168.0.1/24", "10.0.0.1/8"},
			"10.0.100.1",
			true,
		},
	}

	for _, c := range cases {
		ns := NewIPNetSlice(c.inputNets...)
		if ns.Contains(c.inputIP) != c.want {
			t.Errorf("%v.Contains(%s) was wrong: got %t, expected %t", ns, c.inputIP, !c.want, c.want)
		}
	}
}
