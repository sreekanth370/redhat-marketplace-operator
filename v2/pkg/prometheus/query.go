// Copyright 2020 IBM Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"emperror.dev/errors"
	sprig "github.com/Masterminds/sprig/v3"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	marketplacev1beta1 "github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"

	"github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/common"
	"github.com/redhat-marketplace/redhat-marketplace-operator/v2/apis/marketplace/v1beta1"
	"github.com/redhat-marketplace/redhat-marketplace-operator/v2/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = logf.Log.WithName("prometheus")

type PromQueryArgs struct {
	Type          v1beta1.WorkloadType
	MeterDef      types.NamespacedName
	Metric        string
	Query         string
	Start, End    time.Time
	Step          time.Duration
	AggregateFunc string
	GroupBy       []string
	Without       []string

	defaultGroupBy []string
}

type PromQuery struct {
	*PromQueryArgs
}

func NewPromQueryFromLabels(
	meterDefLabels *common.MeterDefPrometheusLabels,
	start, end time.Time,
) *PromQuery {
	workloadType := marketplacev1beta1.WorkloadType(meterDefLabels.WorkloadType)
	duration := time.Hour
	if meterDefLabels.MetricPeriod != nil {
		duration = meterDefLabels.MetricPeriod.Duration
	}

	return NewPromQuery(&PromQueryArgs{
		Metric: meterDefLabels.Metric,
		Type:   workloadType,
		MeterDef: types.NamespacedName{
			Name:      meterDefLabels.MeterDefName,
			Namespace: meterDefLabels.MeterDefNamespace,
		},
		Query:         meterDefLabels.MetricQuery,
		Start:         start,
		End:           end,
		Step:          duration,
		GroupBy:       []string(meterDefLabels.MetricGroupBy),
		Without:       []string(meterDefLabels.MetricWithout),
		AggregateFunc: meterDefLabels.MetricAggregation,
	})

}

func NewPromQuery(
	args *PromQueryArgs,
) *PromQuery {
	pq := &PromQuery{PromQueryArgs: args}
	pq.defaulter()
	return pq
}

func PromQueryFromLabels(
	meterDefLabels *common.MeterDefPrometheusLabels,
	start, end time.Time,
) *PromQuery {

	workloadType := v1beta1.WorkloadType(meterDefLabels.WorkloadType)
	duration := time.Hour
	if meterDefLabels.MetricPeriod != nil {
		duration = meterDefLabels.MetricPeriod.Duration
	}

	return NewPromQuery(&PromQueryArgs{
		Metric: meterDefLabels.Metric,
		Type:   workloadType,
		MeterDef: types.NamespacedName{
			Name:      meterDefLabels.MeterDefName,
			Namespace: meterDefLabels.MeterDefNamespace,
		},
		Query:         meterDefLabels.MetricQuery,
		Start:         start,
		End:           end,
		Step:          duration,
		GroupBy:       []string(meterDefLabels.MetricGroupBy),
		Without:       []string(meterDefLabels.MetricWithout),
		AggregateFunc: meterDefLabels.MetricAggregation,
	})
}

type PrometheusAPI struct {
	v1.API
}

const TypeNotSupportedErr = errors.Sentinel("type is not supported")

func (q *PromQuery) typeNotSupportedError() error {
	err := errors.WithDetails(TypeNotSupportedErr, "type", string(q.Type))
	logger.Error(err, "type not supported", "type", string(q.Type))
	return err
}

func (q *PromQuery) defaulter() {
	q.setDefaultWithout()
	q.setDefaultGroupBy()
}

func dedupeStringSlice(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

func (q *PromQuery) setDefaultWithout() {
	// we want to make sure
	switch q.Type {
	case v1beta1.WorkloadTypePVC:
		q.Without = append(q.Without, "instance", "container", "endpoint", "job", "service", "pod", "pod_uid", "pod_ip")
	case v1beta1.WorkloadTypePod:
		q.Without = append(q.Without, "pod_uid", "pod_ip", "instance", "image_id", "host_ip", "node", "container", "job", "service")
	case v1beta1.WorkloadTypeService:
		q.Without = append(q.Without, "pod_uid", "instance", "container", "endpoint", "job", "pod", "cluster_ip")
	default:
		panic(q.typeNotSupportedError())
	}

	q.Without = append(q.Without, "instance", "container", "endpoint", "job", "cluster_ip")
	q.Without = dedupeStringSlice(q.Without)
}

func (q *PromQuery) setDefaultGroupBy() {
	switch q.Type {
	case v1beta1.WorkloadTypePVC:
		q.defaultGroupBy = []string{"persistentvolumeclaim", "namespace"}
	case v1beta1.WorkloadTypePod:
		q.defaultGroupBy = []string{"pod", "namespace"}
	case v1beta1.WorkloadTypeService:
		q.defaultGroupBy = []string{"service", "namespace"}
	default:
		panic(q.typeNotSupportedError())
	}

	q.defaultGroupBy = dedupeStringSlice(q.defaultGroupBy)
}

const resultQueryTemplateStr = `
{{- .AggregateFunc }} by ({{ default .DefaultGroupBy .GroupBy | join "," }}) (avg({{ .MeterName }}{ {{- .QueryFilters | join "," -}} }) without ({{ .Without | join "," }}) * on({{ .DefaultGroupBy | join "," }}) group_right {{ .Query }}) * on({{ default .DefaultGroupBy .GroupBy | join "," }}) group_right group without({{ .Without | join "," }}) ({{ .Query }})`

var resultQueryTemplate *template.Template = utils.Must(func() (interface{}, error) {
	return template.New("resultQuery").Funcs(sprig.GenericFuncMap()).Parse(resultQueryTemplateStr)
}).(*template.Template)

type ResultQueryArgs struct {
	MeterName, Query, AggregateFunc                string
	QueryFilters, GroupBy, Without, DefaultGroupBy []string
}

func makeLabel(key, value string) string {
	return fmt.Sprintf(`%s="%s"`, key, value)
}

func (q *PromQuery) GetQueryArgs() ResultQueryArgs {
	var meterName string
	queryFilters := []string{
		makeLabel("meter_def_name", q.MeterDef.Name),
		makeLabel("meter_def_namespace", q.MeterDef.Namespace),
	}

	switch q.Type {
	case v1beta1.WorkloadTypePVC:
		meterName = "meterdef_persistentvolumeclaim_info"
		queryFilters = append(queryFilters, makeLabel("phase", "Bound"))
	case v1beta1.WorkloadTypePod:
		meterName = "meterdef_pod_info"
	case v1beta1.WorkloadTypeService:
		meterName = "meterdef_service_info"
	default:
		panic(q.typeNotSupportedError())
	}

	return ResultQueryArgs{
		MeterName:      meterName,
		Query:          q.Query,
		AggregateFunc:  q.AggregateFunc,
		GroupBy:        q.GroupBy,
		QueryFilters:   queryFilters,
		Without:        q.Without,
		DefaultGroupBy: q.defaultGroupBy,
	}
}

func (q *PromQuery) Print() (string, error) {
	var buf strings.Builder
	err := resultQueryTemplate.Execute(&buf, q.GetQueryArgs())
	return buf.String(), err
}

func (p *PrometheusAPI) ReportQuery(query *PromQuery) (model.Value, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timeRange := v1.Range{
		Start: query.Start,
		End:   query.End,
		Step:  query.Step,
	}

	q, err := query.Print()

	if err != nil {
		return nil, nil, err
	}

	logger.Info("executing query", "query", q)

	result, warnings, err := p.QueryRange(ctx, q, timeRange)

	if err != nil {
		logger.Error(err, "querying prometheus", "warnings", warnings)
		return nil, warnings, toError(err)
	}
	if len(warnings) > 0 {
		logger.Info("warnings", "warnings", warnings)
	}

	return result, warnings, nil
}

var ClientError = errors.Sentinel("clientError")
var ClientErrorUnauthorized = errors.Sentinel("clientError: Unauthorized")
var ServerError = errors.Sentinel("serverError")

func toError(err error) error {
	if v, ok := err.(*v1.Error); ok {
		if v.Type == v1.ErrClient {
			if strings.Contains(strings.ToLower(v.Msg), "unauthorized") {
				return errors.Combine(errors.WithStack(ClientErrorUnauthorized), err)
			}

			return errors.Combine(errors.WithStack(ClientError), err)
		}

		return errors.Combine(errors.WithStack(ServerError), err)
	}

	return err
}

type MeterDefinitionQuery struct {
	Start, End time.Time
	Step       time.Duration
}

// Returns a set of elements without duplicates
// Ignore labels such that a pod restart, meterdefinition recreate, or other labels do not generate a new unique element
// Use max over time to get the meter definition most prevalent for the hour
const meterDefinitionQueryStr = `max_over_time(((max without (container, endpoint, instance, job, meter_definition_uid, pod, service) (meterdef_metric_label_info{})) or on() vector(0))[{{ .Step }}:{{ .Step }}])`

var meterDefinitionQueryTemplate *template.Template = utils.Must(func() (interface{}, error) {
	return template.New("meterDefinitionQuery").Funcs(sprig.GenericFuncMap()).Parse(meterDefinitionQueryStr)
}).(*template.Template)

func (q *MeterDefinitionQuery) Print() (string, error) {
	var buf strings.Builder
	err := meterDefinitionQueryTemplate.Execute(&buf, q)
	return buf.String(), err
}

func (p *PrometheusAPI) QueryMeterDefinitions(query *MeterDefinitionQuery) (model.Value, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timeRange := v1.Range{
		Start: query.Start,
		End:   query.End,
		Step:  query.Step,
	}

	q, err := query.Print()
	logger.Info("query params", "query", q, "start", query.Start.Unix(), "end", query.End.Unix(), "step", query.Step.String())

	if err != nil {
		logger.Error(err, "error with query")
		return nil, nil, err
	}

	logger.Info("executing query", "query", q)
	result, warnings, err := p.QueryRange(ctx, q, timeRange)

	if err != nil {
		logger.Error(err, "querying prometheus", "warnings", warnings)
		return nil, warnings, toError(err)
	}
	if len(warnings) > 0 {
		logger.Info("warnings", "warnings", warnings)
	}

	return result, warnings, nil
}
