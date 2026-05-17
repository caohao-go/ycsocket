// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Package metrics defines some common metrics, such as Counter, IGauge, ITimer and IHistogram.
// The method MetricsSink is used to adapt to external monitor systems, such as monitors in our
// company or open source prometheus.
//
// For convenience, we provide two sorts of methods:
// 1. counter
// - reqNumCounter := metrics.Counter("proto.num")
//   reqNumCounter.Incr()
// - metrics.IncrCounter("proto.num", 1)

package metrics

import (
	"fmt"
	"sync"
)

var (
	// allow emit same metrics information to multi external system at the same time.
	metricsSinks    = map[string]Sink{}
	muxMetricsSinks = sync.RWMutex{}

	counters     = map[string]*counter{}
	lockCounters = sync.RWMutex{}
)

// RegisterMetricsSink registers a Sink.
func RegisterMetricsSink(sink Sink) {
	muxMetricsSinks.Lock()
	metricsSinks[sink.Name()] = sink
	muxMetricsSinks.Unlock()
}

// Counter creates a named counter.
func Counter(name string) *counter {
	lockCounters.RLock()
	c, ok := counters[name]
	lockCounters.RUnlock()
	if ok && c != nil {
		return c
	}

	lockCounters.Lock()
	c, ok = counters[name]
	if ok && c != nil {
		lockCounters.Unlock()
		return c
	}
	c = &counter{name: name}
	counters[name] = c
	lockCounters.Unlock()

	return c
}

// IncrCounter increases counter key by value. Counters should accumulate values.
func IncrCounter(key string, value float64) {
	Counter(key).IncrBy(value)
}

// Report reports a multi-dimension record.
func Report(rec Record) (err error) {
	var errs []error
	for _, sink := range metricsSinks {
		err = sink.Report(rec)
		if err != nil {
			errs = append(errs, fmt.Errorf("sink-%s error: %v", sink.Name(), err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("metrics sink error: %v", errs)
}
