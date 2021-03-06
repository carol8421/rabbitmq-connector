// Copyright (c) OpenFaaS Project 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package sdk

import (
	"encoding/json"
	"github.com/openfaas/faas/gateway/requests"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenFaaSClient_buildUrl(t *testing.T) {
	baseUrl := "http://localhost:8080"
	function := "nodeinfo"

	fetchExpected := "http://localhost:8080/system/functions"
	invokeExpected := "http://localhost:8080/function/nodeinfo"

	fetchUrl := buildUrl(baseUrl, "")
	invokeUrl := buildUrl(baseUrl, function)

	if fetchUrl != fetchExpected {
		t.Errorf("Generated URL does not match: Want %s received %s", fetchExpected, fetchUrl)
	}

	if invokeUrl != invokeExpected {
		t.Errorf("Generated URL does not match: Want %s received %s", invokeExpected, invokeUrl)
	}
}

func TestOpenFaaSClient_FetchFunctions(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var functions []requests.Function

		functions = append(functions, requests.Function{
			Name:              "nodeinfo",
			Image:             "functions/nodeinfo",
			InvocationCount:   100,
			Replicas:          2,
			EnvProcess:        "",
			AvailableReplicas: 2,
			Labels:            generateSampleTopics(),
		})
		bytesOut, _ := json.Marshal(functions)
		w.Write(bytesOut)
	}))

	client := OpenFaaSClient{
		url:        mock.URL,
		httpClient: mock.Client(),
	}

	functions, err := client.FetchFunctions()

	if err != nil {
		t.Errorf("Request Failed with %s", err)
	}

	if len(*functions) != 1 {
		t.Errorf("Response is wrong: Want %d received %d", 1, len(*functions))
	}

	for _, function := range *functions {
		labels := *function.Labels
		if labels["topic"] == "" {
			t.Errorf("Response is wrong: Expected label topic to be not empty", 1, len(*functions))
		}
	}

}

func TestOpenFaaSClient_InvokeFunction(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &functionResult{
			Success: true,
			Message: "Send Email transcript",
			Ts:      time.Now().Unix(),
		}

		bytesOut, _ := json.Marshal(response)
		w.Write(bytesOut)
	}))

	client := OpenFaaSClient{
		url:        mock.URL,
		httpClient: mock.Client(),
	}

	rawMessage, err := client.InvokeFunction("transcript", nil)

	if err != nil {
		t.Errorf("Request Failed with %s", err)
	}

	var response functionResult
	err = json.Unmarshal(*rawMessage, &response)

	if err != nil {
		t.Errorf("Parsing Failed with %s", err)
	}

	if response.Success != true {
		t.Errorf("Response is wrong: Want %t received %t", true, response.Success)
	}

	if response.Message != "Send Email transcript" {
		t.Errorf("Response is wrong: Want %s received %s", "Send Email transcript", response.Message)
	}
}

// Util

type functionResult struct {
	Success bool
	Message string
	Ts      int64
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateSampleTopics() *map[string]string {
	labelMap := make(map[string]string)
	labelMap["topic"] = RandStringBytes(10)
	return &labelMap
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
