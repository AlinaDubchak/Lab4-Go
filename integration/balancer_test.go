package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		t.Skip("Integration test is not enabled")
	}

	if !checkBalancer() {
		t.Skip("Balancer is not available")
	}

	// Create a mock server to simulate the backend servers
	mockServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Header().Set("lb-from", "mock-server")
	}))


	resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "mock-server", resp.Header.Get("lb-from"))

	mockServer.Close()
}
func checkBalancer() bool {
	resp, err := client.Get(baseAddress)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func BenchmarkBalancer(b *testing.B) {
	if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
		b.Skip("Integration benchmark is not enabled")
	}

	if !checkBalancer() {
		b.Skip("Balancer is not available")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
		assert.NoError(b, err)
	}
}