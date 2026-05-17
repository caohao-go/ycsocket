// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

type Time time.Time

// UnmarshalJSON Time 类型实现 json unmarshal 方法
func (nt *Time) UnmarshalJSON(data []byte) error {
	tStr := strings.TrimSuffix(strings.TrimPrefix(string(data), `"`), `"`)

	if tStr == "null" {
		return nil
	}

	t, err := ParseTime(tStr, time.Local)
	if err != nil {
		return err
	}
	*nt = Time(t)
	return nil
}

// MarshalJSON Time 类型实现 json marshal 方法
func (nt Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(nt).Format(time.RFC3339Nano) + `"`), nil
}

var typeTimes = []reflect.Type{
	reflect.TypeOf(time.Time{}),
	reflect.TypeOf(Time{}),
}

// RegisterTime 注册新的时间类型，例如上面的 Time，底层必须是 time.Time 类型
func RegisterTime(t reflect.Type) error {
	if t.Kind() != reflect.Struct || !t.ConvertibleTo(typeTimes[0]) {
		return fmt.Errorf("the underlying type of registration must be time")
	}

	typeTimes = append(typeTimes, t)
	return nil
}

// GetRealTime 如果是时间类型，则强制转化为 time.Time 返回。
func GetRealTime(data interface{}) (time.Time, bool) {
	if data == nil {
		return time.Time{}, false
	}

	switch t := data.(type) {
	case time.Time:
		return t, true
	case Time:
		return time.Time(t), true
	case *time.Time:
		return *t, true
	case *Time:
		return time.Time(*t), true
	}

	v := reflect.Indirect(reflect.ValueOf(data))
	if IsTime(v.Type()) {
		// 强制转化为 time.Time 类型
		return v.Convert(typeTimes[0]).Interface().(time.Time), true
	}

	return time.Time{}, false
}

// IsTime 传入类型是否时间
func IsTime(v reflect.Type) bool {
	for _, typeTime := range typeTimes {
		if v == typeTime {
			return true
		}
	}
	return false
}

const dateTimeLen = len(time.DateTime)
const dateOnlyLen = len(time.DateOnly)
const ansicLen = len(time.ANSIC)
const unixDateLen = len(time.UnixDate)
const rfc850Len = len(time.RFC850)
const rfc1123Len = len(time.RFC1123)
const rfc1123zLen = len(time.RFC1123Z)

var commonlyHitErr = errors.New("commonly hit init error")

// ParseTime 解析任何时间
func ParseTime(src interface{}, loc *time.Location, layout ...string) (time.Time, error) {
	if src == nil {
		return time.Time{}, nil
	}

	if loc == nil {
		loc = time.Local
	}

	switch v := src.(type) {
	case []byte:
		return ParseTime(BytesToString(v), loc, layout...)
	case *[]byte:
		return ParseTime(BytesToString(*v), loc, layout...)
	case json.Number:
		s, e := v.Int64()
		if e != nil {
			return time.Time{}, fmt.Errorf("unable to cast %#v of type %T to Time", src, src)
		}

		if s > 17356608000000 { // 微妙
			return time.Unix(s/1e6, (s%1e6)*1e3), nil
		} else if s > 17356608000 { // 毫秒
			return time.Unix(s/1e3, (s%1e3)*1e6), nil
		} else {
			return time.Unix(s, 0), nil
		}
	case int:
		if v > 17356608000 { // 毫秒
			return time.Unix(int64(v/1e3), int64((v%1e3)*1e6)), nil
		} else {
			return time.Unix(int64(v), 0), nil
		}
	case int64:
		if v > 17356608000000 { // 微妙
			return time.Unix(v/1e6, (v%1e6)*1e3), nil
		} else if v > 17356608000 { // 毫秒
			return time.Unix(v/1e3, (v%1e3)*1e6), nil
		} else {
			return time.Unix(v, 0), nil
		}
	case int32:
		return time.Unix(int64(v), 0), nil
	case uint:
		if v > 17356608000 { // 毫秒
			return time.Unix(int64(v/1e3), int64((v%1e3)*1e6)), nil
		} else {
			return time.Unix(int64(v), 0), nil
		}
	case uint64:
		if v > 17356608000000 { // 微妙
			return time.Unix(int64(v/1e6), int64((v%1e6)*1e3)), nil
		} else if v > 17356608000 { // 毫秒
			return time.Unix(int64(v/1e3), int64((v%1e3)*1e6)), nil
		} else {
			return time.Unix(int64(v), 0), nil
		}
	case uint32:
		return time.Unix(int64(v), 0), nil
	case string:
		l := len(v)
		if l == 0 {
			return time.Time{}, nil
		}

		if len(layout) > 0 && layout[0] != "" {
			return time.ParseInLocation(layout[0], v, loc)
		}

		i, err := strconv.ParseUint(v, 10, 64)
		if err == nil && i > 631123200 { //时间戳，且大于 1990-01-01 00:00:00
			return ParseTime(i, loc)
		}

		var ct time.Time
		var ce = commonlyHitErr

		switch l {
		case dateTimeLen:
			ct, ce = time.ParseInLocation(time.DateTime, v, loc)
		case dateOnlyLen:
			ct, ce = time.ParseInLocation(time.DateOnly, v, loc)
		default:
			if l > 19 {
				switch v[:3] {
				case "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun":
					switch l {
					case ansicLen - 1, ansicLen:
						ct, ce = time.ParseInLocation(time.ANSIC, v, loc)
					case unixDateLen - 1, unixDateLen:
						ct, ce = time.ParseInLocation(time.UnixDate, v, loc)
					case rfc850Len:
						ct, ce = time.ParseInLocation(time.RFC850, v, loc)
					case rfc1123Len:
						ct, ce = time.ParseInLocation(time.RFC1123, v, loc)
					case rfc1123zLen:
						ct, ce = time.ParseInLocation(time.RFC1123Z, v, loc)
					}
				default:
					ct, ce = time.ParseInLocation(time.RFC3339Nano, v, loc) // 首先校验是否最常用的 RFC3339Nano 格式，json.Marshal 一般是这个格式。
				}
			}
		}

		if ce == nil {
			return ct, nil
		}

		return dateparse.ParseIn(v, loc)
	default:
		t, ok := GetRealTime(src)
		if ok {
			return t, nil
		}

		return time.Time{}, fmt.Errorf("unable to cast %#v of type %T to Time", src, src)
	}
}

// Now 根据 orm type 获取当前时间
func Now(typ Type) interface{} {
	switch typ {
	case TypeInt, TypeInt32, TypeInt64,
		TypeUint, TypeUint32, TypeUint64:
		return time.Now().Unix()
	case TypeString:
		return time.Now().Format(time.RFC3339Nano)
	default:
		return time.Now()
	}
}

// GetFormatTime 获取格式化时间
func GetFormatTime(data interface{}, timeFmt string) interface{} {
	if timeFmt != "" {
		t, ok := GetRealTime(data)
		if ok {
			return t.Format(timeFmt)
		}
	}

	return data
}

func Day(d int) time.Duration {
	return 24 * time.Duration(d) * time.Hour
}

func Hour(d int) time.Duration {
	return time.Duration(d) * time.Hour
}

func Second(d int) time.Duration {
	return time.Duration(d) * time.Second
}

// Millisecond 返回毫秒 time.Duration
func Millisecond(d int) time.Duration {
	return time.Duration(d) * time.Millisecond
}

// MSecTime 返回当前毫秒级时间戳
func MSecTime() int64 {
	return time.Now().UnixMilli() / 1000
}
