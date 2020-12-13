package pkg

import "time"

// FormatDatetime 安装 format 格式分析 timestr
func FormatDatetime(timestr string, format string) (time.Time, error) {
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
