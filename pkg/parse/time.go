package parse

import "time"

// Str2Time 根据 `format` 分析 time
func Str2Time(timestr string, format string) (time.Time, error) {
	var (
		t   time.Time
		err error
	)
	t, err = time.Parse(format, timestr)
	if err != nil {
		return t, err
	}
	return t, nil
}
