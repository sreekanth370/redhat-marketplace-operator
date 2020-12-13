package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	marketplacev1alpha1 "github.com/redhat-marketplace/redhat-marketplace-operator/pkg/apis/marketplace/v1alpha1"

	// . "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// . "github.com/onsi/gomega/gstruct"
)

// RoundTripFunc is a type that represents a round trip function call for std http lib
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip is a wrapper function that calls an external function for mocking
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func GetTestAPI(trip RoundTripFunc) v1.API {
	conf := api.Config{
		Address:      "http://localhost:9090",
		RoundTripper: trip,
	}
	client, err := api.NewClient(conf)

	Expect(err).To(Succeed())

	v1api := v1.NewAPI(client)
	return v1api
}

func GenerateMeterInfoResponse(meterdefinitions []marketplacev1alpha1.MeterDefinition) []byte {
	results := []map[string]interface{}{}
	for _, mdef := range meterdefinitions {
		labels := mdef.ToPrometheusLabels()

		for _, labelMap := range labels {
			labelMap["name"] = mdef.Name
			labelMap["namespace"] = mdef.Namespace
			results = append(results, map[string]interface{}{
				"metric": labelMap,
				"values": [][]interface{}{
					{1, "1"},
					{2, "1"},
				},
			})
		}
	}

	data := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "matrix",
			"result":     results,
		},
	}

	bytes, _ := json.Marshal(&data)

	fmt.Println(string(bytes))

	return bytes
}

func MockResponseRoundTripper(file string, meterdefinitions []marketplacev1alpha1.MeterDefinition) RoundTripFunc {
	return func(req *http.Request) *http.Response {
		headers := make(http.Header)
		headers.Add("content-type", "application/json")

		Expect(req.URL.String()).To(Equal("http://localhost:9090/api/v1/query_range"), "url does not match expected")

		fileBytes, err := ioutil.ReadFile(file)

		Expect(err).To(Succeed(), "failed to load mock file for response")
		defer req.Body.Close()
		body, err := ioutil.ReadAll(req.Body)

		Expect(err).To(Succeed())

		query, _ := url.ParseQuery(string(body))

		if strings.Contains(query["query"][0], "meterdef_metric_label_info{}") {
			fmt.Println("using meter_label_info")
			meterDefInfo := GenerateMeterInfoResponse(meterdefinitions)
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBuffer(meterDefInfo)),
				// Must be set to non-nil value or it panics
				Header: headers,
			}
		}

		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBuffer(fileBytes)),
			// Must be set to non-nil value or it panics
			Header: headers,
		}
	}
}