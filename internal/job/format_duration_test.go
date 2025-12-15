package job

import (
	"testing"
	"time"
)

func TestFormatDuration_Extended(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		out  string
	}{
		// Zero / negative
		{"zero", 0, "now"},
		{"negative", -1 * time.Hour, "now"},

		// Seconds → minutes rounding
		{"1 second", 1 * time.Second, "less than a minute"},
		{"29 seconds", 29 * time.Second, "less than a minute"},
		{"30 seconds", 30 * time.Second, "1 minute"},
		{"89 seconds", 89 * time.Second, "1 minute"},
		{"90 seconds", 90 * time.Second, "2 minutes"},

		// Pure minutes
		{"1 minute", 1 * time.Minute, "1 minute"},
		{"2 minutes", 2 * time.Minute, "2 minutes"},
		{"59 minutes", 59 * time.Minute, "59 minutes"},

		// Minutes → hours boundary
		{"60 minutes", 60 * time.Minute, "1 hour 0 minutes"},
		{"61 minutes", 61 * time.Minute, "1 hour 1 minute"},
		{"119 minutes", 119 * time.Minute, "1 hour 59 minutes"},
		{"120 minutes", 120 * time.Minute, "2 hours 0 minutes"},

		// Mixed hours/minutes
		{"3h 1m", 3*time.Hour + 1*time.Minute, "3 hours 1 minute"},
		{"3h 59m", 3*time.Hour + 59*time.Minute, "3 hours 59 minutes"},
		{"3h 59m 29s", 3*time.Hour + 59*time.Minute + 29*time.Second, "3 hours 59 minutes"},
		{"3h 59m 31s", 3*time.Hour + 59*time.Minute + 31*time.Second, "4 hours 0 minutes"},

		// Pure hours
		{"1 hour", 1 * time.Hour, "1 hour 0 minutes"},
		{"23 hours", 23 * time.Hour, "23 hours 0 minutes"},

		// Hours → days boundary
		{"24 hours", 24 * time.Hour, "1 day 0 hours"},
		{"25 hours", 25 * time.Hour, "1 day 1 hour"},
		{"47 hours", 47 * time.Hour, "1 day 23 hours"},
		{"48 hours", 48 * time.Hour, "2 days 0 hours"},

		// Rounding into next day
		{"23h 59m 29s", 23*time.Hour + 59*time.Minute + 29*time.Second, "23 hours 59 minutes"},
		{"23h 59m 31s", 23*time.Hour + 59*time.Minute + 31*time.Second, "1 day 0 hours"},

		// Multi-day durations
		{"2 days", 2 * 24 * time.Hour, "2 days 0 hours"},
		{"2d 5h", 2*24*time.Hour + 5*time.Hour, "2 days 5 hours"},
		{"7 days", 7 * 24 * time.Hour, "7 days 0 hours"},
		{"10d 12h", 10*24*time.Hour + 12*time.Hour, "10 days 12 hours"},

		// Large durations
		{"30 days", 30 * 24 * time.Hour, "30 days 0 hours"},
		{"45d 6h", 45*24*time.Hour + 6*time.Hour, "45 days 6 hours"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.in)

			t.Logf("input: %v → output: %q", tt.in, got)

			if got != tt.out {
				t.Fatalf(
					"formatDuration(%v) = %q, want %q",
					tt.in, got, tt.out,
				)
			}
		})
	}
}
