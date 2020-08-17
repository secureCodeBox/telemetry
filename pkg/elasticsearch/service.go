package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
)

// Service can be used to interact with elasticsearch
type Service interface {
	Create(index string, document interface{}) error
}

// ServiceClient can be used to interact with elasticsearch
type ServiceClient struct {
	client *elasticsearch.Client
}

// Setup Initializes a elasticsearch Client
func Setup() (*ServiceClient, error) {
	elasticsearchURL, _ := os.LookupEnv("ELASTIC_URL")
	username, _ := os.LookupEnv("ELASTIC_USERNAME")
	password, _ := os.LookupEnv("ELASTIC_PASSWORD")
	cfg := elasticsearch.Config{
		Addresses: []string{
			elasticsearchURL,
		},
		Username: username,
		Password: password,
	}
	client, err := elasticsearch.NewClient(cfg)

	if err != nil {
		return nil, err
	}

	return &ServiceClient{client: client}, nil
}

// Create pushes a new document to the specified index
func (service ServiceClient) Create(index string, document interface{}) error {
	bytes, err := json.Marshal(document)
	if err != nil {
		return errors.New("Failed to encode document to json")
	}

	response, err := service.client.Index("telemetry-test", strings.NewReader(string(bytes)))

	if err != nil {
		return err
	}

	if response.IsError() {
		return fmt.Errorf("Elasticsearch request failed with status code: '%d'", response.StatusCode)
	}

	return nil
}
