package gohelp

import "testing"

func TestParseHex(t *testing.T) {
	for _, in := range []string{"#5fafff", "5fafff"} {
		got, err := parseHex(in)
		if err != nil || got != Hex(in) {
			t.Errorf("parseHex(%q) = %q, %v; want %q", in, got, err, Hex(in))
		}
	}
	for _, in := range []string{"", "#5faff", "#0000000", "#gggggg"} {
		t.Run("bad "+in, func(t *testing.T) {
			if _, err := parseHex(in); err == nil {
				t.Errorf("parseHex(%q) returned no error", in)
			}
			defer func() {
				if recover() == nil {
					t.Errorf("Hex(%q) did not panic", in)
				}
			}()
			Hex(in)
		})
	}
}
