package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestBalancer() {
	addr1 := getServerIndexByAddress("192.168.0.0:80")
	addr2 := getServerIndexByAddress("127.0.0.0:8080")
	addr3 := getServerIndexByAddress("26.143.218.9:80")
	
	assert.Equal(s.T(), 0, addr1)
	assert.Equal(s.T(), 2, addr2)
	assert.Equal(s.T(), 1, addr3)
}

func (s *TestSuite) TestServersHealth() {
	result := make([]string, len(serversPool))

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	parsedURL1, _ := url.Parse(server1.URL)
	hostURL1 := parsedURL1.Host

	parsedURL2, _ := url.Parse(server1.URL)
	hostURL2 := parsedURL2.Host

	servers := []string{
		hostURL1,
		hostURL2,
		"server3:8080",
	}

	checkServersHealth(servers, result)
	time.Sleep(12 * time.Second)

	assert.Equal(s.T(), hostURL1, result[0])
	assert.Equal(s.T(), hostURL2, result[1])
	assert.Equal(s.T(), "", result[2])
}