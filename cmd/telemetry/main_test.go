package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type Service struct {
	mock.Mock
}

func (service *Service) Create(index string, document interface{}) error {
	args := service.Called(index, document)
	return args.Error(0)
}

func encode(data interface{}) io.Reader {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic("Failed to encode document to json")
	}

	return strings.NewReader(string(bytes))
}

func TestShouldCreateDocumentWhenDataIsGoodAndElasticsearchWorks(t *testing.T) {
	testObj := new(Service)
	testObj.On("Create", mock.AnythingOfType("string"), mock.AnythingOfType("TelemetryDataDocument")).Return(nil)

	router := setupRouter(testObj)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		"/v1/submit",
		encode(
			TelemetryData{
				Version:            "v2.0.42",
				InstalledScanTypes: []string{"nmap", "sslyze"},
			},
		),
	)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestShouldRejectDataWithInproperScanTypes(t *testing.T) {
	testObj := new(Service)
	router := setupRouter(testObj)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		"/v1/submit",
		encode(
			TelemetryData{
				Version:            "v2.0.42",
				InstalledScanTypes: []string{"fooooobarrrrrrrrr"},
			},
		),
	)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestShouldIndicateErrorWhenElasticsearchThrownsError(t *testing.T) {
	testObj := new(Service)
	testObj.On(
		"Create",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("TelemetryDataDocument"),
	).Return(
		fmt.Errorf("Elasticsearch request failed with status code: '%d'", 401),
	)

	router := setupRouter(testObj)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(
		"POST",
		"/v1/submit",
		encode(
			TelemetryData{
				Version:            "v2.0.42",
				InstalledScanTypes: []string{"nmap", "sslyze"},
			},
		),
	)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}
