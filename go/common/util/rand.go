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

// Package util provides public goroutine-safe random function.
// The implementation is similar to grpc random functions. Additionally,
// the seed function is provided to be called from the outside, and
// the random functions are provided as a body's methods.
package util

import (
	"math/rand"
	"sync"
)

type SafeRand struct {
	r  *rand.Rand
	mu sync.Mutex
}

func NewSafeRand(seed int64) *SafeRand {
	c := &SafeRand{
		r: rand.New(rand.NewSource(seed)),
	}
	return c
}

func (c *SafeRand) Intn(n int) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	res := c.r.Intn(n)
	return res
}
