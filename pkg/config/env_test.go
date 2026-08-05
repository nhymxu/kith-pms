package config

import "testing"

func TestEffectiveMaxUploadBytes(t *testing.T) {
	cases := []struct {
		mb   int
		want int64
	}{
		{0, 32 * 1024 * 1024},
		{-1, 32 * 1024 * 1024},
		{1, 1 * 1024 * 1024},
		{64, 64 * 1024 * 1024},
	}

	for _, tc := range cases {
		c := Config{MaxUploadSizeMB: tc.mb}
		if got := c.EffectiveMaxUploadBytes(); got != tc.want {
			t.Errorf("MaxUploadSizeMB=%d: got %d, want %d", tc.mb, got, tc.want)
		}
	}
}

func TestEffectiveImageMaxEdgePX(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, 1600},
		{-100, 1600},
		{512, 512},
		{4000, 4000},
	}

	for _, tc := range cases {
		c := Config{ImageMaxEdgePX: tc.in}
		if got := c.EffectiveImageMaxEdgePX(); got != tc.want {
			t.Errorf("ImageMaxEdgePX=%d: got %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestEffectiveImageJPEGQuality(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, 85},
		{-5, 85},
		{1, 1},
		{60, 60},
		{100, 100},
		{101, 100},
		{1000, 100},
	}

	for _, tc := range cases {
		c := Config{ImageJPEGQuality: tc.in}
		if got := c.EffectiveImageJPEGQuality(); got != tc.want {
			t.Errorf("ImageJPEGQuality=%d: got %d, want %d", tc.in, got, tc.want)
		}
	}
}

// configDefaults must agree with the accessor fallbacks; a divergence would mean
// a value set via Load behaves differently from one built directly.
func TestConfigDefaultsMatchFallbacks(t *testing.T) {
	cases := map[string]int{
		"MAX_UPLOAD_SIZE_MB": defaultMaxUploadSizeMB,
		"IMAGE_MAX_EDGE_PX":  defaultImageMaxEdgePX,
		"IMAGE_JPEG_QUALITY": defaultImageJPEGQuality,
	}

	for key, want := range cases {
		got, ok := configDefaults[key].(int)
		if !ok {
			t.Errorf("configDefaults[%q] missing or not an int", key)

			continue
		}

		if got != want {
			t.Errorf("configDefaults[%q] = %d, want %d", key, got, want)
		}
	}
}
