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

package metrics

// counter defines the counter. counter is report to each external Sink-able system.
type counter struct {
	name string
}

// Incr increases counter by one.
func (c *counter) Incr() {
	c.IncrBy(1)
}

// IncrBy increases counter by v and reports for each external Sink-able systems.
func (c *counter) IncrBy(v float64) {
	rec := NewSingleDimensionMetrics(c.name, v, PolicySUM)
	for _, sink := range metricsSinks {
		sink.Report(rec)
	}
}
