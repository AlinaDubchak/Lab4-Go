package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BalancerSuite struct{
	suite.Suite
}

func (s *BalancerSuite) TestBalancer() {
	addr1 := GetServerIndexByAddress("192.168.0.0:80")
	addr2 := GetServerIndexByAddress("127.0.0.0:8080")
	addr3 := GetServerIndexByAddress("26.143.218.9:80")
	
	assert.Equal(s.T(), 0, addr1)
	assert.Equal(s.T(), 2, addr2)
	assert.Equal(s.T(), 1, addr3)
}

func MockHealth(dst string) bool {
	if dst == "server1:8080" {
		return true
	} else if dst == "server2:8080" {
		return false
	}
	return false
}

func MockHealthAllTrue(dst string) bool {
	return true
}

func (s *BalancerSuite) TestServersHealth() {
	healthChecker := &HealthChecker{}
    healthChecker.health = MockHealth
}

func TestHealthChecker(t *testing.T) {
		healthChecker := &HealthChecker{}
		healthChecker.serverHealthStatus = map[string]bool{}
		healthChecker.health = MockHealth
	
		healthChecker.CheckAllServers()
		assert.Equal(t, map[string]bool{"server1:8080": true, "server2:8080": false, "server3:8080": false}, healthChecker.serverHealthStatus)
	
		healthyServers := healthChecker.GetHealthyServers()
		assert.Equal(t, []string{"server1:8080"}, healthyServers)
	
		healthChecker.health = MockHealthAllTrue
		healthChecker.CheckAllServers()
		healthyServers = healthChecker.GetHealthyServers()
		assert.Equal(t, []string{"server1:8080", "server2:8080", "server3:8080"}, healthyServers)
}