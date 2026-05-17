package util

import (
	"strconv"
	"time"
)

// ========================= 辅助函数 =========================

func DateYmd() string     { return time.Now().Format("20060102") }
func LastDateYmd() string { return time.Now().Add(-24 * time.Hour).Format("20060102") }
func NextDateYmd() string { return time.Now().Add(24 * time.Hour).Format("20060102") }
func DateW() string       { _, w := time.Now().ISOWeek(); return strconv.Itoa(w) }
func LastDateW() string {
	_, w := time.Now().Add(-7 * 24 * time.Hour).ISOWeek()
	return strconv.Itoa(w)
}
func NextDateW() string { _, w := time.Now().Add(7 * 24 * time.Hour).ISOWeek(); return strconv.Itoa(w) }
func DateM() string     { return time.Now().Format("200601") }

// ========================= 时间辅助 =========================

// LeftTimeToTomorrow 返回距明天0点的秒数
func LeftTimeToTomorrow() int {
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return int(tomorrow.Sub(now).Seconds())
}

// GetNextMonday 获取下周一日期 Ymd
func GetNextMonday() string {
	now := time.Now()
	w := int(now.Weekday())
	d := 1
	if w == 0 {
		d = 1
	} else {
		d = 8 - w
	}
	return now.AddDate(0, 0, d).Format("20060102")
}

// LeftTimeToNextMonday 返回距下周一0点的秒数
func LeftTimeToNextMonday() int {
	t, _ := time.ParseInLocation("20060102", GetNextMonday(), time.Now().Location())
	return int(t.Sub(time.Now()).Seconds())
}

// LeftTimeToWeekend2100 返回距本周日21点的秒数
func LeftTimeToWeekend2100() int {
	now := time.Now()
	t := now
	for t.Weekday() != time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	target := time.Date(t.Year(), t.Month(), t.Day(), 21, 0, 0, 0, t.Location())
	left := int(target.Sub(now).Seconds())
	if left < 0 {
		return 0
	}
	return left
}

// DianCoinRefreshTime 点金手刷新倒计时
func DianCoinRefreshTime() int {
	now := time.Now()
	his := now.Hour()*10000 + now.Minute()*100 + now.Second()
	today := now.Format("2006-01-02")
	if his > 0 && his < 90000 {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", today+" 09:00:00", now.Location())
		return int(t.Sub(now).Seconds())
	} else if his > 90000 && his < 190000 {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", today+" 19:00:00", now.Location())
		return int(t.Sub(now).Seconds())
	}
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", today+" 23:59:59", now.Location())
	return int(t.Sub(now).Seconds())
}
