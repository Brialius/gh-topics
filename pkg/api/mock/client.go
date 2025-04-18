// Package mock provides mock implementations for testing the API package
package mock

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cli/go-gh/v2/pkg/api"
)

// MockRESTClient is a mock implementation of the REST client used in the API package
type MockRESTClient struct {
	// GetResponses maps endpoint paths to mock responses
	GetResponses map[string]interface{}
	// PutResponses maps endpoint paths to success/failure status
	PutResponses map[string]error
	// Calls records the calls made to the client for verification
	Calls []Call
}

// Call represents a recorded API call
type Call struct {
	Method   string
	Endpoint string
	Body     interface{}
}

// NewMockRESTClient creates a new mock REST client
func NewMockRESTClient() *MockRESTClient {
	return &MockRESTClient{
		GetResponses: make(map[string]interface{}),
		PutResponses: make(map[string]error),
		Calls:        []Call{},
	}
}

// Get implements the REST client Get method
func (m *MockRESTClient) Get(path string, response interface{}) error {
	m.Calls = append(m.Calls, Call{
		Method:   "GET",
		Endpoint: path,
	})

	mockResp, ok := m.GetResponses[path]
	if !ok {
		return fmt.Errorf("mock response not found for path: %s", path)
	}

	// Convert the mock response to the expected response type
	jsonData, err := json.Marshal(mockResp)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, response)
}

// Put implements the REST client Put method
func (m *MockRESTClient) Put(path string, body io.Reader, response interface{}) error {
	// Read and store the request body
	var requestBody interface{}
	if body != nil {
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			return err
		}

		m.Calls = append(m.Calls, Call{
			Method:   "PUT",
			Endpoint: path,
			Body:     requestBody,
		})
	} else {
		m.Calls = append(m.Calls, Call{
			Method:   "PUT",
			Endpoint: path,
		})
	}

	// Return the configured response for this path
	return m.PutResponses[path]
}

// MockGraphQLClient is a mock implementation of the GraphQL client used in the API package
type MockGraphQLClient struct {
	// QueryResponses maps query names to mock responses
	QueryResponses map[string]interface{}
	// Calls records the calls made to the client for verification
	Calls []GraphQLCall
}

// GraphQLCall represents a recorded GraphQL API call
type GraphQLCall struct {
	QueryName string
	Variables map[string]interface{}
}

// NewMockGraphQLClient creates a new mock GraphQL client
func NewMockGraphQLClient() *MockGraphQLClient {
	return &MockGraphQLClient{
		QueryResponses: make(map[string]interface{}),
		Calls:          []GraphQLCall{},
	}
}

// Query implements the GraphQL client Query method
func (m *MockGraphQLClient) Query(name string, query interface{}, variables map[string]interface{}) error {
	m.Calls = append(m.Calls, GraphQLCall{
		QueryName: name,
		Variables: variables,
	})

	mockResp, ok := m.QueryResponses[name]
	if !ok {
		return fmt.Errorf("mock response not found for query: %s", name)
	}

	// Convert the mock response to the expected query type
	jsonData, err := json.Marshal(mockResp)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, query)
}

// MockClientFactory is a function that creates mock clients for testing
type MockClientFactory struct {
	RESTClient    *MockRESTClient
	GraphQLClient *MockGraphQLClient
}

// NewRESTClient returns a mock REST client
func (f *MockClientFactory) NewRESTClient(opts api.ClientOptions) (*MockRESTClient, error) {
	return f.RESTClient, nil
}

// NewGraphQLClient returns a mock GraphQL client
func (f *MockClientFactory) NewGraphQLClient(opts api.ClientOptions) (*MockGraphQLClient, error) {
	return f.GraphQLClient, nil
}
