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

package url

import (
	"net/url"
)

func ParseQuery(s string) (map[string]string, error) {
	ret := map[string]string{}

	if s == "" {
		return ret, nil
	}

	tmp, err := url.ParseQuery(s)
	if err != nil {
		return nil, err
	}

	for k, v := range tmp {
		ret[k] = v[0]
	}

	return ret, nil
}

func ParamEncode(params map[string]string) string {
	value := url.Values{}
	for k, v := range params {
		value[k] = []string{v}
	}

	return value.Encode()
}
