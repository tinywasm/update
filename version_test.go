package update

import "testing"

func TestIsOutdated(t *testing.T) {
	cases := []struct {
		cur, lat string
		want     bool
	}{
		{"v0.1.0", "v0.2.0", true},
		{"0.2.0", "v0.2.0", false},
		{"v1.2.3", "v1.2.2", false},
		{"dev", "v9.9.9", false}, // never bother "dev" builds
		{"v0.2.0", "garbage", false},
	}
	for _, c := range cases {
		if got := IsOutdated(c.cur, c.lat); got != c.want {
			t.Errorf("IsOutdated(%q,%q)=%v want %v", c.cur, c.lat, got, c.want)
		}
	}
}
