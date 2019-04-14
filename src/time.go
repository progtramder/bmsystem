package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func timeString() string {
	t := time.Now()
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}

//yyyy-mm-dd hh:mm
func parseTime(tm string) (t time.Time, err error) {
	if tm == "" {
		err = errors.New("不能为空")
		return
	}
	match, _ := regexp.MatchString(`^\d{4}-\d{1,2}-\d{1,2} \d{1,2}:\d{1,2}$`, tm)
	if !match {
		err = errors.New("时间格式错误")
		return
	}

	timeString := strings.Split(tm, " ")
	ymdString := timeString[0]
	hmString := timeString[1]

	ymd := strings.Split(ymdString, "-")
	year, _ := strconv.ParseInt(ymd[0], 10, 32)
	month, _ := strconv.ParseInt(ymd[1], 10, 32)
	day, _ := strconv.ParseInt(ymd[2], 10, 32)
	hm := strings.Split(hmString, ":")
	hour, _ := strconv.ParseInt(hm[0], 10, 32)
	minute, _ := strconv.ParseInt(hm[1], 10, 32)

	local := time.Now().Location()
	t = time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), 0, 0, local)
	return
}
