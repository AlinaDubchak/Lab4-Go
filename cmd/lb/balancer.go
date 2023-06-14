package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/AlinaDubchak/Lab4-Go/httptools"
	"github.com/AlinaDubchak/Lab4-Go/signal"
)

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}
	activeServers = make([]string, len(serversPool))
	healthChecker = &HealthChecker{}
)

type HealthChecker struct {
	serverHealthStatus map[string]bool
	health             func(dst string) bool
	mu             sync.Mutex
}

func (hc *HealthChecker) CheckAllServers() {
	for _, server := range serversPool {
		if hc.health(server) {
			hc.serverHealthStatus[server] = true
		} else {
			hc.serverHealthStatus[server] = false
		}
	}
}

func (hc *HealthChecker) GetHealthyServers() []string {
	var healthyServers []string
	for _, server := range serversPool {
		if hc.serverHealthStatus[server] {
			healthyServers = append(healthyServers, server)
		}
	}
	return healthyServers
}

func Scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func Health(dst string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", Scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func Forward(dst string, rw http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = dst
	fwdRequest.URL.Scheme = Scheme()
	fwdRequest.Host = dst

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", dst)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

func main() {
	
	flag.Parse()
	CheckServersHealth(serversPool, activeServers)

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		healthChecker.mu.Lock()
		server := activeServers[len(serversPool)]
		healthChecker.mu.Unlock()
		_ =Forward(server, rw, r)
		Forward(server, rw, r)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}

func GetServerIndexByAddress(addr string) int {
	hashed := EncryptAddress(addr)
	serverIndex := int(hashed) % len(activeServers)
	return serverIndex
}

func EncryptAddress(addr string) uint32 {
	hash := fnv.New32()
	_, err := hash.Write([]byte(addr))
	if err != nil {
		log.Printf("Failed to compute hash: %s", err)
		return 0
	}
	return hash.Sum32()
}

func CheckServersHealth(servers []string, result []string) {
	for i, server := range servers {
		StartHealthMonitoring(server, i, result)
	}
}

func StartHealthMonitoring(server string, index int, result []string) {
	go func() {
		for range time.Tick(10 * time.Second) {
			if Health(server) {
				result[index] = server
			}
			healthChecker.mu.Lock()
			activeServers = activeServers[:0]
			for _, value := range serversPool {
				if value != "" {
					activeServers = append(activeServers, value)
				}
			}
			healthChecker.mu.Unlock()
			log.Println(server, Health(server))
		}
	}()
}
