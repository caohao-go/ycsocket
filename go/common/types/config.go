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
	"reflect"
	"strings"
	"time"

	"server_golang/common/errs"
)

type ConfType uint8

const (
	ConfTypeBool         ConfType = 1  // bool
	ConfTypeString       ConfType = 2  // string
	ConfTypeInt          ConfType = 3  // int
	ConfTypeUint         ConfType = 4  // uint
	ConfTypeFloat        ConfType = 5  // float
	ConfTypeEnum         ConfType = 6  // 枚举
	ConfTypeTime         ConfType = 7  // 时间
	ConfTypeArray        ConfType = 8  // array
	ConfTypeMap          ConfType = 9  // map
	ConfTypeMulti        ConfType = 10 // multiple
	ConfTypeTimeInterval ConfType = 11 // 时间区间
)

type Config map[string]interface{}

// GetBool 获取 bool 类型配置
func (f Config) GetBool(key string) (ret bool, exist bool) {
	return GetBool(f, key)
}

// GetString 获取 string 类型配置
func (f Config) GetString(key string) (ret string, exist bool) {
	return GetString(f, key)
}

// GetInt 获取 int 类型配置
func (f Config) GetInt(key string) (ret int64, exist bool, err error) {
	ret, exist, err = GetInt64(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get int config error: %v", err)
	}
	return
}

// GetUint 获取 uint 类型配置
func (f Config) GetUint(key string) (ret uint64, exist bool, err error) {
	ret, exist, err = GetUint64(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get uint config error: %v", err)
	}
	return
}

// GetFloat 获取 float 类型配置
func (f Config) GetFloat(key string) (ret float64, exist bool, err error) {
	ret, exist, err = GetFloat64(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get float config error: %v", err)
	}
	return
}

// GetBytes 获取 bytes 类型配置
func (f Config) GetBytes(key string) (ret []byte, exist bool) {
	return GetBytes(f, key)
}

// GetEnum 获取枚举（单选）类型配置
func (f Config) GetEnum(key string) (ret string, exist bool) {
	return GetString(f, key)
}

// GetTime 获取 date、time 类型配置
func (f Config) GetTime(key string, loc ...*time.Location) (ret time.Time, exist bool, err error) {
	if len(f) == 0 {
		return time.Time{}, false, nil
	}

	value, ok := f[key]
	if !ok {
		return time.Time{}, false, nil
	}

	if value == nil {
		return time.Time{}, false, nil
	}

	l := time.Local
	if len(loc) > 0 {
		l = loc[0]
	}

	t, err := ParseTime(value, l)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get time config error: %v", err)
	}

	return t, true, nil
}

// GetTimeInterval 获取 date、time 时间区间
func (f Config) GetTimeInterval(key string, loc ...*time.Location) (start, end time.Time, exist bool, err error) {
	if len(f) == 0 {
		return time.Time{}, time.Time{}, false, nil
	}

	value, ok := f[key]
	if !ok {
		return time.Time{}, time.Time{}, false, nil
	}

	if value == nil {
		return time.Time{}, time.Time{}, false, nil
	}

	times, ok := value.([]time.Time)
	if ok && len(times) == 2 {
		return times[0], times[1], true, nil
	}

	var startTimeStr interface{}
	var endTimeStr interface{}

	if IsArray(value) {
		valArr, _ := ToArray(value)
		if len(valArr) != 2 {
			return time.Time{}, time.Time{}, true, errs.New(errs.ErrConfig,
				"get date interval config error: config value should have start time and end time")
		}

		startTimeStr = valArr[0]
		endTimeStr = valArr[1]
	} else {
		str := ToString(value)
		t := strings.Split(str, "~")
		if len(t) != 2 {
			return time.Time{}, time.Time{}, true, errs.New(errs.ErrConfig,
				"get date interval config error: config value should have start time and end time")
		}

		startTimeStr = t[0]
		endTimeStr = t[1]
	}

	l := time.Local
	if len(loc) > 0 {
		l = loc[0]
	}

	start, err = ParseTime(startTimeStr, l)
	if err != nil {
		return time.Time{}, time.Time{}, true,
			errs.Newf(errs.ErrConfig, "get time interval config start time error: %v", err)
	}

	end, err = ParseTime(endTimeStr, l)
	if err != nil {
		return time.Time{}, time.Time{}, true,
			errs.Newf(errs.ErrConfig, "get time interval config end time error: %v", err)
	}

	return start, end, true, nil
}

// GetStringArray 获取 string 数组配置
func (f Config) GetStringArray(key string) (ret []string, exist bool, err error) {
	ret, exist, err = GetStringArray(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get string array config error: %v", err)
	}
	return
}

// GetIntArray 获取 int 数组配置
func (f Config) GetIntArray(key string) (ret []int64, exist bool, err error) {
	ret, exist, err = GetInt64Array(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get int array config error: %v", err)
	}
	return
}

// GetUintArray 获取 uint 数组配置
func (f Config) GetUintArray(key string) (ret []uint64, exist bool, err error) {
	ret, exist, err = GetUint64Array(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get uint array config error: %v", err)
	}
	return
}

// GetFloatArray 获取 float 数组配置
func (f Config) GetFloatArray(key string) (ret []float64, exist bool, err error) {
	ret, exist, err = GetFloat64Array(f, key)
	if err != nil {
		err = errs.Newf(errs.ErrConfig, "get float array config error: %v", err)
	}
	return
}

// GetMapConf 获取 map 类型配置
func (f Config) GetMapConf(key string) (Config, bool, error) {
	tmp, exist, err := GetMap(f, key)
	if err != nil {
		return nil, exist, errs.Newf(errs.ErrConfig, "get map config error: %v", err)
	}

	return Config(tmp), exist, nil
}

// GetMultiConf 获取配置数组
func (f Config) GetMultiConf(key string) (ret []Config, exist bool, err error) {
	value, ok := f[key]
	if !ok {
		return nil, false, nil
	}

	if value == nil {
		return nil, true, nil
	}

	switch arrVal := value.(type) {
	case []Config:
		return arrVal, true, nil
	case []interface{}:
		ret = make([]Config, len(arrVal))
		for k, arrItem := range arrVal {
			im, e := ToMap(arrItem, "")
			if e != nil {
				return nil, true, errs.Newf(errs.ErrConfig, "get multiple conf error: %v", e)
			}
			ret[k] = Config(im)
		}
	case []map[string]interface{}:
		ret = make([]Config, len(arrVal))
		for k, arrItem := range arrVal {
			ret[k] = arrItem
		}
	default:
		v := reflect.ValueOf(value)
		if IsNil(v) {
			return nil, true, nil
		}

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if !IsArray(v) {
			return nil, true, errs.Newf(errs.ErrConfig, "get multiple conf error: value is not array")
		}

		l := v.Len()
		ret = make([]Config, l)

		for i := 0; i < l; i++ {
			im, e := ToMap(Interface(v.Index(i)), "")
			if e != nil {
				return nil, true, errs.Newf(errs.ErrConfig, "get multiple conf error: %v", e)
			}
			ret[i] = Config(im)
		}
	}

	return ret, true, nil
}
