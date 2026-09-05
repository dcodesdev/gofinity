package main

import "testing"

func TestConstants(t *testing.T) {
	if AppName != "Gofinity" {
		t.Errorf("AppName = %q, want %q", AppName, "Gofinity")
	}
	if MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want %d", MaxRetries, 3)
	}
}

func TestRetryStatus(t *testing.T) {
	tests := []struct {
		name string
		used int
		want string
	}{
		{
			name: "none used",
			used: 0,
			want: "Gofinity: 0 of 3 retries used, 3 remaining",
		},
		{
			name: "some used",
			used: 2,
			want: "Gofinity: 2 of 3 retries used, 1 remaining",
		},
		{
			name: "all used",
			used: 3,
			want: "Gofinity: 3 of 3 retries used, 0 remaining",
		},
		{
			name: "more than allowed is clamped",
			used: 9,
			want: "Gofinity: 3 of 3 retries used, 0 remaining",
		},
		{
			name: "negative is clamped",
			used: -4,
			want: "Gofinity: 0 of 3 retries used, 3 remaining",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RetryStatus(tt.used); got != tt.want {
				t.Errorf("RetryStatus(%d) = %q, want %q", tt.used, got, tt.want)
			}
		})
	}
}

func TestRemaining(t *testing.T) {
	tests := []struct {
		used int
		want int
	}{
		{used: -1, want: 3},
		{used: 0, want: 3},
		{used: 1, want: 2},
		{used: 3, want: 0},
		{used: 8, want: 0},
	}

	for _, tt := range tests {
		if got := Remaining(tt.used); got != tt.want {
			t.Errorf("Remaining(%d) = %d, want %d", tt.used, got, tt.want)
		}
	}
}
