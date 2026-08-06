package smb

import "time"

const filetimeEpochOffset = 116444736000000000

func timeToFiletime(t time.Time) uint64 {
	return uint64(t.UnixNano()/100 + filetimeEpochOffset)
}

func filetimeToTime(ft uint64) time.Time {
	return time.Unix(0, (int64(ft)-filetimeEpochOffset)*100).UTC()
}
