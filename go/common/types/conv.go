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
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"git.code.oa.com/pcg-csd/trpc-ext/orm"
	"github.com/spf13/cast"
)

// ToString 接口转字符串
func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case *string:
		return *v
	case orm.NullString:
		return string(v)
	case *orm.NullString:
		return string(*v)
	case []byte:
		return string(v)
	case *[]byte:
		return string(*v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	default:
		return cast.ToString(value)
	}
}

// ToBytes 接口转字节码
func ToBytes(value interface{}) []byte {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return v
	case *[]byte:
		return *v
	default:
		str := ToString(value)
		return []byte(str)
	}
}

func ToBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case *bool:
		return *v
	case []byte:
		b, _ := strconv.ParseBool(BytesToString(v))
		return b
	case *[]byte:
		b, _ := strconv.ParseBool(BytesToString(*v))
		return b
	default:
		return cast.ToBool(value)
	}
}

// ToInt64E 将 interface{} 转为 int64
func ToInt64E(v interface{}) int64 {
	ret, _ := ToInt64(v)
	return ret
}

func ToInt64(value interface{}) (int64, error) {
	if value == nil {
		return 0, nil
	}
	switch v := value.(type) {
	case int64:
		return v, nil
	case *int64:
		return *v, nil
	case []byte:
		return strconv.ParseInt(BytesToString(v), 10, 64)
	case *[]byte:
		return strconv.ParseInt(BytesToString(*v), 10, 64)
	default:
		return cast.ToInt64E(value)
	}
}

func ToUint64(value interface{}) (uint64, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case uint64:
		return v, nil
	case *uint64:
		return *v, nil
	case []byte:
		return ToUint64(BytesToString(v))
	case *[]byte:
		return ToUint64(BytesToString(*v))
	default:
		return cast.ToUint64E(value)
	}
}

// ToFloat64E 将 interface{} 转为 float64
func ToFloat64E(v interface{}) float64 {
	ret, _ := ToFloat64(v)
	return ret
}

func ToFloat64(value interface{}) (float64, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case *float64:
		return *v, nil
	case []byte:
		return ToFloat64(BytesToString(v))
	case *[]byte:
		return ToFloat64(BytesToString(*v))
	default:
		return cast.ToFloat64E(value)
	}
}

// ToIntE 将 interface{} 转为 int
func ToIntE(v interface{}) int {
	ret, _ := ToInt(v)
	return ret
}

func ToInt(value interface{}) (int, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case *int:
		return *v, nil
	case []byte:
		return ToInt(BytesToString(v))
	case *[]byte:
		return ToInt(BytesToString(*v))
	default:
		return cast.ToIntE(value)
	}
}

func ToInt8(value interface{}) (int8, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case int8:
		return v, nil
	case *int8:
		return *v, nil
	case []byte:
		return ToInt8(BytesToString(v))
	case *[]byte:
		return ToInt8(BytesToString(*v))
	default:
		return cast.ToInt8E(value)
	}
}

func ToInt16(value interface{}) (int16, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case int16:
		return v, nil
	case *int16:
		return *v, nil
	case []byte:
		return ToInt16(BytesToString(v))
	case *[]byte:
		return ToInt16(BytesToString(*v))
	default:
		return cast.ToInt16E(value)
	}
}

func ToInt32(value interface{}) (int32, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case int32:
		return v, nil
	case *int32:
		return *v, nil
	case []byte:
		return ToInt32(BytesToString(v))
	case *[]byte:
		return ToInt32(BytesToString(*v))
	default:
		return cast.ToInt32E(value)
	}
}

func ToUint(value interface{}) (uint, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case uint:
		return v, nil
	case *uint:
		return *v, nil
	case []byte:
		return ToUint(BytesToString(v))
	case *[]byte:
		return ToUint(BytesToString(*v))
	default:
		return cast.ToUintE(value)
	}
}

func ToUint8(value interface{}) (uint8, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case uint8:
		return v, nil
	case *uint8:
		return *v, nil
	case []byte:
		return ToUint8(BytesToString(v))
	case *[]byte:
		return ToUint8(BytesToString(*v))
	default:
		return cast.ToUint8E(value)
	}
}

func ToUint16(value interface{}) (uint16, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case uint16:
		return v, nil
	case *uint16:
		return *v, nil
	case []byte:
		return ToUint16(BytesToString(v))
	case *[]byte:
		return ToUint16(BytesToString(*v))
	default:
		return cast.ToUint16E(value)
	}
}

func ToUint32(value interface{}) (uint32, error) {
	if value == nil {
		return 0, nil
	}

	switch v := value.(type) {
	case uint32:
		return v, nil
	case *uint32:
		return *v, nil
	case []byte:
		return ToUint32(BytesToString(v))
	case *[]byte:
		return ToUint32(BytesToString(*v))
	default:
		return cast.ToUint32E(value)
	}
}

func ToArrayE(value interface{}) []interface{} {
	ret, _ := ToArray(value)
	return ret
}

// ToArray 接口转数组
func ToArray(value interface{}) (ret []interface{}, err error) {
	if value == nil {
		return []interface{}{}, nil
	}

	switch val := value.(type) {
	case []interface{}:
		return val, nil
	case []map[string]interface{}:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []Map:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []string:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []int:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []int8:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []int16:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []int32:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []int64:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []uint:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []uint16:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []uint32:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []uint64:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []float32:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case []float64:
		ret = make([]interface{}, len(val))
		for k, v := range val {
			ret[k] = v
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return []interface{}{}, errors.New("value is not array")
	case string:
		ret = []interface{}{}
		err = json.Unmarshal([]byte(val), &ret)
		return
	case []byte:
		ret = []interface{}{}
		err = json.Unmarshal(val, &ret)
		return
	default:
		v := reflect.Indirect(reflect.ValueOf(value))
		if !IsArray(v) {
			return []interface{}{}, errors.New("value is not array")
		}

		l := v.Len()
		ret = make([]interface{}, l)

		for i := 0; i < l; i++ {
			ret[i] = Interface(v.Index(i))
		}
	}

	return
}

func ToMapE(value interface{}, op ...int8) Map {
	v, _ := ToMap(value, "", op...)
	return v
}

// ToMap 接口转 map
func ToMap(value interface{}, tag string, op ...int8) (Map, error) {
	if value == nil {
		return Map{}, nil
	}

	switch v := value.(type) {
	case Map:
		return v, nil
	case *Map:
		return *v, nil
	case map[string]interface{}:
		return v, nil
	case *map[string]interface{}:
		return *v, nil
	case map[string]string:
		tmp := make(map[string]interface{}, len(v))
		for k, v := range v {
			tmp[k] = v
		}
		return tmp, nil
	case *map[string]string:
		tmp := make(map[string]interface{}, len(*v))
		for k, v := range *v {
			tmp[k] = v
		}
		return tmp, nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return Map{}, errors.New("value is not map")
	case []byte:
		ret := Map{}
		err := json.Unmarshal(v, &ret)
		return ret, err
	case string:
		ret := Map{}
		err := json.Unmarshal([]byte(v), &ret)
		return ret, err
	case *[]byte:
		ret := Map{}
		err := json.Unmarshal(*v, &ret)
		return ret, err
	case *string:
		ret := Map{}
		err := json.Unmarshal([]byte(*v), &ret)
		return ret, err
	case orm.NullString:
		ret := Map{}
		err := json.Unmarshal([]byte(v), &ret)
		return ret, err
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsStruct(rv.Type()) {
		return StructToMap(rv, tag, op...), nil
	}

	if rv.Kind() != reflect.Map {
		return Map{}, errors.New("value is not map")
	}

	// 将非 string 的 key 转化为 string
	ret := make(map[string]interface{}, rv.Len())
	for _, k := range rv.MapKeys() {
		ret[ToString(Interface(k))] = Interface(rv.MapIndex(k))
	}

	return ret, nil
}

// ToMapString 接口转 map
func ToMapString(value interface{}, tag string, op ...int8) (map[string]string, error) {
	v, err := ToMap(value, tag, op...)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]string, len(v))
	for k, v := range v {
		ret[k] = ToString(v)
	}
	return ret, nil
}

func ToMapArrayE(value interface{}) []Map {
	ret, _ := ToMapArray(value, "")
	return ret
}

// ToMapArray 接口转map数组
func ToMapArray(value interface{}, tag string, op ...int8) (ret []Map, err error) {
	if value == nil {
		return []Map{}, nil
	}

	switch v := value.(type) {
	case []Map:
		return v, nil
	case []map[string]interface{}:
		ret = make([]Map, len(v))
		for i, m := range v {
			ret[i] = m
		}
		return ret, nil
	case *[]map[string]interface{}:
		ret = make([]Map, len(*v))
		for i, m := range *v {
			ret[i] = m
		}
		return ret, nil
	case []map[string]string:
		ret = make([]Map, len(v))
		for i, m := range v {
			tmp, _ := ToMap(m, tag, op...)
			ret[i] = tmp
		}
		return ret, nil
	case *[]map[string]string:
		ret = make([]Map, len(*v))
		for i, m := range *v {
			tmp, _ := ToMap(m, tag, op...)
			ret[i] = tmp
		}
		return ret, nil
	case []byte:
		ret = []Map{}
		err = json.Unmarshal(v, &ret)
		return ret, err
	case string:
		ret = []Map{}
		err = json.Unmarshal([]byte(v), &ret)
		return ret, err
	case *[]byte:
		ret = []Map{}
		err = json.Unmarshal(*v, &ret)
		return ret, err
	case *string:
		ret = []Map{}
		err = json.Unmarshal([]byte(*v), &ret)
		return ret, err
	case orm.NullString:
		ret = []Map{}
		err = json.Unmarshal([]byte(v), &ret)
		return ret, err
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsStructArray(rv) {
		v := StructsToMaps(rv, tag, op...)
		ret = make([]Map, len(v))
		for i, m := range v {
			ret[i] = m
		}
		return ret, nil
	}

	arr, err := ToArray(Indirect(value))
	if err != nil {
		return []Map{}, errors.New("value is not map array")
	}

	ret = make([]Map, len(arr))
	for i, item := range arr {
		m, err := ToMap(item, tag, op...)
		if err != nil {
			return []Map{}, errors.New("value is not map array")
		}
		ret[i] = m
	}

	return ret, nil
}

// ToStringArray 接口转字符串数组
func ToStringArray(value interface{}) ([]string, error) {
	if value == nil {
		return []string{}, nil
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []int:
		ret := make([]string, len(v))
		for k, vv := range v {
			ret[k] = ToString(vv)
		}
		return ret, nil
	case []int64:
		ret := make([]string, len(v))
		for k, vv := range v {
			ret[k] = ToString(vv)
		}
		return ret, nil
	case []interface{}:
		ret := make([]string, len(v))
		for k, item := range v {
			ret[k] = ToString(item)
		}
		return ret, nil
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsArray(rv) {
		l := rv.Len()
		ret := make([]string, l)

		for i := 0; i < l; i++ {
			ret[i] = ToString(Interface(rv.Index(i)))
		}

		return ret, nil
	}

	return []string{}, errors.New("value is not array")
}

// ToInt64Array 接口转int64数组
func ToInt64Array(value interface{}) (ret []int64, err error) {
	if value == nil {
		return []int64{}, nil
	}

	switch v := value.(type) {
	case []int64:
		return v, nil
	case []interface{}:
		ret = make([]int64, len(v))
		for k, item := range v {
			ret[k], err = ToInt64(item)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsArray(rv) {
		l := rv.Len()
		ret = make([]int64, l)

		for i := 0; i < l; i++ {
			ret[i], err = ToInt64(Interface(rv.Index(i)))
			if err != nil {
				return []int64{}, err
			}
		}

		return ret, nil
	}

	return []int64{}, errors.New("value is not array")
}

// ToIntArray 接口转 int 数组
func ToIntArray(value interface{}) (ret []int, err error) {
	if value == nil {
		return []int{}, nil
	}

	switch v := value.(type) {
	case []int:
		return v, nil
	case []interface{}:
		ret = make([]int, len(v))
		for k, item := range v {
			ret[k], err = ToInt(item)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	case []int64:
		ret = make([]int, len(v))
		for k, item := range v {
			ret[k] = int(item)
		}
		return ret, nil
	case []float64:
		ret = make([]int, len(v))
		for k, item := range v {
			ret[k] = int(item)
		}
		return ret, nil
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsArray(rv) {
		l := rv.Len()
		ret = make([]int, l)

		for i := 0; i < l; i++ {
			ret[i], err = ToInt(Interface(rv.Index(i)))
			if err != nil {
				return []int{}, err
			}
		}

		return ret, nil
	}

	return []int{}, errors.New("value is not array")
}

// ToUint64Array 接口转 uint64 数组
func ToUint64Array(value interface{}) (ret []uint64, err error) {
	if value == nil {
		return []uint64{}, nil
	}

	switch v := value.(type) {
	case []uint64:
		return v, nil
	case []interface{}:
		ret = make([]uint64, len(v))
		for k, item := range v {
			ret[k], err = ToUint64(item)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsArray(rv) {
		l := rv.Len()
		ret = make([]uint64, l)

		for i := 0; i < l; i++ {
			ret[i], err = ToUint64(Interface(rv.Index(i)))
			if err != nil {
				return []uint64{}, err
			}
		}

		return ret, nil
	}

	return []uint64{}, errors.New("value is not array")
}

// ToFloat64Array 接口转 float64 数组
func ToFloat64Array(value interface{}) (ret []float64, err error) {
	if value == nil {
		return []float64{}, nil
	}

	switch v := value.(type) {
	case []float64:
		return v, nil
	case []interface{}:
		ret = make([]float64, len(v))
		for k, item := range v {
			ret[k], err = ToFloat64(item)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if IsArray(rv) {
		l := rv.Len()
		ret = make([]float64, l)

		for i := 0; i < l; i++ {
			ret[i], err = ToFloat64(Interface(rv.Index(i)))
			if err != nil {
				return []float64{}, err
			}
		}

		return ret, nil
	}

	return []float64{}, errors.New("value is not array")
}

// ToIntMap 接口转 map[int]Map（key 为 int 的 map）
func ToIntMap(value interface{}) (map[int]Map, error) {
	if value == nil {
		return map[int]Map{}, nil
	}

	switch v := value.(type) {
	case map[int]Map:
		return v, nil
	case *map[int]Map:
		return *v, nil
	case map[int]map[string]interface{}:
		ret := make(map[int]Map, len(v))
		for k, m := range v {
			ret[k] = m
		}
		return ret, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		string, bool, []byte, time.Time:
		return map[int]Map{}, errors.New("value is not int-keyed map")
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if rv.Kind() != reflect.Map {
		return map[int]Map{}, errors.New("value is not map")
	}

	ret := make(map[int]Map, rv.Len())
	for _, k := range rv.MapKeys() {
		keyInt, err := ToInt(Interface(k))
		if err != nil {
			// key 无法转为 int，用 ToString 兜底转 hash 或跳过
			continue
		}
		val := Interface(rv.MapIndex(k))
		if subMap, ok := val.(Map); ok {
			ret[keyInt] = subMap
		} else if subM, err2 := ToMap(val, ""); err2 == nil {
			ret[keyInt] = subM
		}
	}

	return ret, nil
}

// ToIntMapInt64 接口转 map[int]int64（key 为 int，value 为 int64 的 map）
func ToIntMapInt64(value interface{}) (map[int]int64, error) {
	if value == nil {
		return map[int]int64{}, nil
	}

	switch v := value.(type) {
	case map[int]int64:
		return v, nil
	case *map[int]int64:
		return *v, nil
	case map[int]map[string]interface{}:
		ret := make(map[int]int64, len(v))
		for k, m := range v {
			if id, ok := m["id"]; ok {
				if i, err := ToInt64(id); err == nil {
					ret[k] = i
				} else if i, ok := id.(int64); ok {
					ret[k] = i
				}
			}
		}
		return ret, nil
	case map[int]interface{}:
		ret := make(map[int]int64, len(v))
		for k, val := range v {
			if i, err := ToInt64(val); err == nil {
				ret[k] = i
			}
		}
		return ret, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		string, bool, []byte, time.Time:
		return map[int]int64{}, errors.New("value is not int-keyed map")
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if rv.Kind() != reflect.Map {
		return map[int]int64{}, errors.New("value is not map")
	}

	ret := make(map[int]int64, rv.Len())
	for _, k := range rv.MapKeys() {
		keyInt, err := ToInt(Interface(k))
		if err != nil {
			continue
		}
		val := Interface(rv.MapIndex(k))
		if i, err := ToInt64(val); err == nil {
			ret[keyInt] = i
		}
	}

	return ret, nil
}

// ToIntMapInt 接口转 map[int]int（key 为 int，value 为 int 的 map）
func ToIntMapInt(value interface{}) (map[int]int, error) {
	if value == nil {
		return map[int]int{}, nil
	}

	switch v := value.(type) {
	case map[int]int:
		return v, nil
	case *map[int]int:
		return *v, nil
	case map[int]map[string]interface{}:
		ret := make(map[int]int, len(v))
		for k, m := range v {
			if id, ok := m["id"]; ok {
				if i, err := ToInt(id); err == nil {
					ret[k] = i
				} else if i, ok := id.(int); ok {
					ret[k] = i
				}
			}
		}
		return ret, nil
	case map[int]interface{}:
		ret := make(map[int]int, len(v))
		for k, val := range v {
			if i, err := ToInt(val); err == nil {
				ret[k] = i
			}
		}
		return ret, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		string, bool, []byte, time.Time:
		return map[int]int{}, errors.New("value is not int-keyed map")
	}

	rv := reflect.Indirect(reflect.ValueOf(value))
	if rv.Kind() != reflect.Map {
		return map[int]int{}, errors.New("value is not map")
	}

	ret := make(map[int]int, rv.Len())
	for _, k := range rv.MapKeys() {
		keyInt, err := ToInt(Interface(k))
		if err != nil {
			continue
		}
		val := Interface(rv.MapIndex(k))
		if i, err := ToInt(val); err == nil {
			ret[keyInt] = i
		}
	}

	return ret, nil
}

// ObjectToMap 将结构体转换为 map
func ObjectToMap(source interface{}) Map {
	codec := Map2Structure{
		tagName:    "json",
		squash:     true,
		weaklyType: true,
		l:          time.Local,
	}

	ret := Map{}
	codec.Decode(source, &ret)
	return ret
}

// ObjectsToMaps 将结构体转换为 []Map
func ObjectsToMaps(source interface{}) []Map {
	codec := Map2Structure{
		tagName:    "json",
		squash:     true,
		weaklyType: true,
		l:          time.Local,
	}

	ret := []Map{}
	codec.Decode(source, &ret)
	return ret
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Note it may break if the implementation of string or slice header changes in the future go versions.
func BytesToString(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes converts string to a byte slice without memory allocation.
//
// Note it may break if the implementation of string or slice header changes in the future go versions.
func StringToBytes(s string) (b []byte) {
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))

	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}
