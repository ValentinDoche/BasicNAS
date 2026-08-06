package smb

import (
	"testing"
	"time"
)

func TestFiletimeRoundTrip(t *testing.T) {
	want := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	got := filetimeToTime(timeToFiletime(want))
	if !got.Equal(want) {
		t.Errorf("filetimeToTime(%v) = %v, want %v", want, got, want)
	}
}
