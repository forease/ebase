/**
 * Ebase frame for daemon program
 * Author Jonsen Yang
 * Date 2013-07-05
 * Copyright (c) 2013 ForEase Times Technology Co., Ltd. All rights reserved.
 */
package ebase

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

// encode md5
func Md5(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

// string to uint
func GetStrUint(d string) uint {
	a, _ := strconv.Atoi(d)
	return uint(a)
}

// string to int
func GetStrInt(d string) int {
	a, _ := strconv.Atoi(d)
	return a
}

// string to uint64
func GetStrUint64(d string) uint64 {
	if d == "" {
		return uint64(0)
	}

	ret, err := strconv.ParseUint(d, 10, 64)

	if err != nil {
		//log.Println("getStrUint64 err", err)
	}

	return ret
}

// string to int64
func GetStrInt64(d string) int64 {
	if d == "" {
		return int64(0)
	}

	ret, err := strconv.ParseInt(d, 10, 64)

	if err != nil {
		//log.Println("getStrUint64 err", err)
	}

	return ret
}

// string to float64
func GetStrFloat64(s string) float64 {
	if s == "" {
		return float64(0)
	}

	ret, err := strconv.ParseFloat(s, 64)

	if err != nil {
		//log.Println("getStrUint64 err", err)
	}

	return ret

}

// string to uint32
func GetStrUint32(d string) uint32 {
	a, _ := strconv.Atoi(d)
	return uint32(a)
}

// uint64 to string
func GetUint64Str(d uint64) string {
	return strconv.FormatUint(d, 10)
}

// int64 to string
func GetInt64Str(d int64) string {
	return strconv.FormatInt(d, 10)
}

// uint to string
func GetUintStr(d uint) string {
	return strconv.Itoa(int(d))
}

// int to string
func GetIntStr(d int) string {
	return strconv.Itoa(d)
}

func GetTimeAgo(t int64) (s string) {
	tt := time.Now().Unix() - t

	if tt < 60 {
		s = GetInt64Str(tt) + " 秒以前"
	} else if tt < 3600 {
		m := tt / 60
		s = GetInt64Str(m) + " 分钟以前"
	} else if tt < 86400 {
		m := tt / 3600
		s = GetInt64Str(m) + " 小时以前"
	} else if tt < 2592000 {
		m := tt / 86400
		s = GetInt64Str(m) + " 天以前"
	} else if tt < 2592000*12 {
		m := tt / 86400 * 30
		s = GetInt64Str(m) + " 月以前"
	} else {
		m := tt / 2592000 * 12
		s = GetInt64Str(m) + " 月以前"
	}

	return
}

// int slice to string
func GetIntArrToStr(i []int, d string) (s string) {
	var ss []string
	for _, v := range i {
		ss = append(ss, GetIntStr(v))
	}

	return strings.Join(ss, d)
}
