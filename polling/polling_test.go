package polling

import (
	"testing"
)

func TestFormatDowntime(t *testing.T) {
	tests := []struct {
		sec      int64
		expected string
	}{
		{0, "0s"},
		{15, "15s"},
		{59, "59s"},
		{60, "1m"},
		{320, "5m 20s"},
		{3600, "1h"},
		{3660, "1h 1m"},
		{86400, "1d"},
		{90000, "1d 1h"},
	}

	for _, tt := range tests {
		got := formatDowntime(tt.sec)
		if got != tt.expected {
			t.Errorf("formatDowntime(%d) = %q; want %q", tt.sec, got, tt.expected)
		}
	}
}
