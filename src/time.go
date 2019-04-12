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

func parseTime(tm string) (t time.Time, err error) {
	match, _ := regexp.MatchString(`^\d+-\d+-\d+$`, tm)
	if !match {
		err = errors.New("时间格式错误")
		return
	}

	timeString := strings.Split(tm, "-")
	year, _ := strconv.ParseInt(timeString[0], 10, 32)
	month, _ := strconv.ParseInt(timeString[1], 10, 32)
	day, _ := strconv.ParseInt(timeString[2], 10, 32)
	local := time.Now().Location()
	t = time.Date(int(year), time.Month(month), int(day), 7, 59, 59, 0, local)
	return
}
