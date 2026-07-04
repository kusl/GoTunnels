package captcha

import (
	"testing"

	"github.com/kusl/GoTunnels/internal/store"
)

func TestClampSync(t *testing.T) {
	cases := []struct {
		name string
		in   store.CaptchaSyncInput
		want store.CaptchaSyncInput
	}{
		{
			name: "passthrough",
			in:   store.CaptchaSyncInput{ManualDelta: 3, AutoDelta: 40, CurrentStreak: 7, BestStreak: 12},
			want: store.CaptchaSyncInput{ManualDelta: 3, AutoDelta: 40, CurrentStreak: 7, BestStreak: 12},
		},
		{
			name: "negatives floor to zero",
			in:   store.CaptchaSyncInput{ManualDelta: -1, AutoDelta: -99, CurrentStreak: -5, BestStreak: -5},
			want: store.CaptchaSyncInput{},
		},
		{
			name: "deltas capped",
			in:   store.CaptchaSyncInput{ManualDelta: maxDeltaPerSync + 1, AutoDelta: 1 << 60},
			want: store.CaptchaSyncInput{ManualDelta: maxDeltaPerSync, AutoDelta: maxDeltaPerSync},
		},
		{
			name: "streaks capped",
			in:   store.CaptchaSyncInput{CurrentStreak: maxStreak + 5, BestStreak: maxStreak + 5},
			want: store.CaptchaSyncInput{CurrentStreak: maxStreak, BestStreak: maxStreak},
		},
		{
			name: "best raised to at least current",
			in:   store.CaptchaSyncInput{CurrentStreak: 9, BestStreak: 2},
			want: store.CaptchaSyncInput{CurrentStreak: 9, BestStreak: 9},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampSync(tc.in); got != tc.want {
				t.Fatalf("clampSync(%+v) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	if got := clamp(5, 0, 10); got != 5 {
		t.Fatalf("clamp mid = %d", got)
	}
	if got := clamp(-5, 0, 10); got != 0 {
		t.Fatalf("clamp low = %d", got)
	}
	if got := clamp(50, 0, 10); got != 10 {
		t.Fatalf("clamp high = %d", got)
	}
}
