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
	"time"

	"github.com/mitchellh/mapstructure"
)

type Map2Structure struct {
	tagName    string
	weaklyType bool
	squash     bool
	l          *time.Location
}

var typeString = reflect.TypeOf("")
var typeTime1 = reflect.TypeOf(time.Time{})
var typeTime2 = reflect.TypeOf(Time{})

func (m *Map2Structure) Decode(src, dest interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: m.weaklyType,
		Squash:           m.squash,
		Result:           dest,
		TagName:          m.tagName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			func(str reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if str == typeString && IsTime(t) {
					tt, err := ParseTime(data.(string), m.l)
					if err != nil {
						return nil, err
					}

					switch t {
					case typeTime1:
						return tt, nil
					case typeTime2:
						return Time(tt), nil
					default:
						return reflect.ValueOf(tt).Convert(t).Interface(), nil //强制转化为接收类型
					}
				}

				return data, nil
			},
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}
