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

// Policy is the metrics aggregation policy.
type Policy int

// All available Policy(s).
const (
	PolicyNONE  = 0 // Undefined
	PolicySET   = 1 // instantaneous value
	PolicySUM   = 2 // summary
	PolicyAVG   = 3 // average
	PolicyMAX   = 4 // maximum
	PolicyMIN   = 5 // minimum
	PolicyMID   = 6 // median
	PolicyTimer = 7 // timer
)

// Sink defines the interface an external monitor system should provide.
type Sink struct{}

// Name returns noop.
func (n *Sink) Name() string {
	return "noop"
}

// Report does nothing.
func (n *Sink) Report(rec Record) error {
	return nil
}

// Record is the single record.
//
// terminologies:
//   - dimension name
//     is an attribute of a body, often used to filter body, such as a photo album business model
//     includes region and server room.
//   - dimension value
//     refines the dimension. For example, the regions of the album business model include Shenzhen,
//     Shanghai, etc., the region is the dimension, and Shenzhen and Shanghai are the dimension
//     values.
//   - metric
//     is a measurement, used to aggregate and calculate. For example, request count of album business
//     model in ShenZhen Telecom is a metric.
type Record struct {
	Name string // the name of the record
	// dimension name: such as region, host and disk number.
	// dimension value: such as region=ShangHai.
	dimensions []*Dimension
	metrics    []*Metrics
}

// Dimension defines the dimension.
type Dimension struct {
	Name  string
	Value string
}

// GetName returns the record name.
func (r *Record) GetName() string {
	return r.Name
}

// GetDimensions returns dimensions.
func (r *Record) GetDimensions() []*Dimension {
	if r == nil {
		return nil
	}
	return r.dimensions
}

// GetMetrics returns metrics.
func (r *Record) GetMetrics() []*Metrics {
	if r == nil {
		return nil
	}
	return r.metrics
}

// NewSingleDimensionMetrics creates a Record with no dimension and only one metric.
func NewSingleDimensionMetrics(name string, value float64, policy Policy) Record {
	r := Record{
		dimensions: nil,
		metrics: []*Metrics{
			{name: name, value: value, policy: policy},
		},
	}
	return r
}

// ReportSingleDimensionMetrics creates and reports a Record with no dimension and only one metric.
func ReportSingleDimensionMetrics(name string, value float64, policy Policy) error {
	r := Record{
		dimensions: nil,
		metrics: []*Metrics{
			{name: name, value: value, policy: policy},
		},
	}
	return Report(r)
}

// NewMultiDimensionMetrics creates a Record with multiple dimensions and metrics.
// Deprecated use NewMultiDimensionMetricsX instead.
func NewMultiDimensionMetrics(dimensions []*Dimension, metrics []*Metrics) Record {
	return NewMultiDimensionMetricsX("", dimensions, metrics)
}

// NewMultiDimensionMetricsX creates a named Record with multiple dimensions and metrics.
func NewMultiDimensionMetricsX(name string, dimensions []*Dimension, metrics []*Metrics) Record {
	r := Record{
		Name:       name,
		dimensions: dimensions,
		metrics:    metrics,
	}
	return r
}

// ReportMultiDimensionMetrics creates and reports a Record with multiple dimensions and metrics.
// Deprecated use ReportMultiDimensionMetricsX instead.
func ReportMultiDimensionMetrics(dimensions []*Dimension, metrics []*Metrics) error {
	return ReportMultiDimensionMetricsX("", dimensions, metrics)
}

// ReportMultiDimensionMetricsX creates and reports a named Record with multiple dimensions and
// metrics.
func ReportMultiDimensionMetricsX(name string, dimensions []*Dimension, metrics []*Metrics) error {
	r := Record{
		Name:       name,
		dimensions: dimensions,
		metrics:    metrics,
	}
	return Report(r)
}

// Metrics defines the metric.
type Metrics struct {
	name   string  // metric name
	value  float64 // metric value
	policy Policy  // aggregation policy
}

// NewMetrics creates a new Metrics.
func NewMetrics(name string, value float64, policy Policy) *Metrics {
	return &Metrics{name, value, policy}
}

// Name returns the metrics name.
func (m *Metrics) Name() string {
	if m == nil {
		return ""
	}
	return m.name
}

// Value returns the metrics value.
func (m *Metrics) Value() float64 {
	if m == nil {
		return 0
	}
	return m.value
}

// Policy returns the metrics policy.
func (m *Metrics) Policy() Policy {
	if m == nil {
		return PolicyNONE
	}
	return m.policy
}
