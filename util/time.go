package util

import "time"

// 获取当前时间离24:00的过期时间
func GetExpiredByDay() (expire time.Duration) {
	t0 := time.Now()
	tm1 := time.Date(t0.Year(), t0.Month(), t0.Day(), 0, 0, 0, 0, t0.Location())
	tm2 := tm1.AddDate(0, 0, 0)
	expired := int(60*60*24 - (t0.Unix() - tm2.Unix()))
	expire = time.Duration(expired) * time.Second
	return expire
}

// 获取当月第一天时间戳
func GetMonthFirstDayTimeStamp() (t int64) {
	t0 := time.Now()
	tm1 := time.Date(t0.Year(), t0.Month(), t0.Day(), 0, 0, 0, 0, t0.Location())
	tm2 := tm1.AddDate(0, 0, 0)
	t = tm2.AddDate(0, 0, -tm2.Day()+1).Unix()
	return t
}
