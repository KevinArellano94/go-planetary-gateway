package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestConcurrentHeavyRequests(t *testing.T) {
	// The handler for the heavy task
	heavyHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println("Starting heavy task")
		time.Sleep(2 * time.Second) // Reduced sleep time for faster tests
		fmt.Fprintf(w, "Heavy task complete!")
		log.Println("Finished heavy task")
	}

	server := httptest.NewServer(http.HandlerFunc(heavyHandler))
	defer server.Close()

	startTime := time.Now()

	var wg sync.WaitGroup
	numRequests := 100
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}

	wg.Wait()

	duration := time.Since(startTime)

	// If requests were sequential, it would take > 6s (3 * 2s).
	// Concurrently, it should be a bit over 2s.
	if duration > 3*time.Second {
		t.Errorf("Expected concurrent requests to take ~2 seconds, but it took %v", duration)
	}
}

func TestConcurrentPlanetCreation(t *testing.T) {
	// Reset state for this test
	mutex.Lock()
	planets = []Planet{}
	nextID = 1
	mutex.Unlock()

	// Setup the router from main.go
	mux := http.NewServeMux()
	mux.HandleFunc("/planets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPlanets(w, r)
		case http.MethodPost:
			createPlanet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	var wg sync.WaitGroup
	numRequests := 100
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func(i int) {
			defer wg.Done()
			planetName := fmt.Sprintf("Planet-%d", i)
			planet := Planet{Name: planetName}
			body, _ := json.Marshal(planet)
			resp, err := http.Post(server.URL+"/planets", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			resp.Body.Close()
		}(i)
	}

	wg.Wait()

	// Verify the final state
	resp, err := http.Get(server.URL + "/planets")
	if err != nil {
		t.Fatalf("Failed to get planets: %v", err)
	}
	defer resp.Body.Close()

	var finalPlanets []Planet
	if err := json.NewDecoder(resp.Body).Decode(&finalPlanets); err != nil {
		t.Fatalf("Failed to decode planets: %v", err)
	}

	if len(finalPlanets) != numRequests {
		t.Errorf("Expected %d planets, but got %d", numRequests, len(finalPlanets))
	}

	// Check for duplicate IDs
	idSet := make(map[int]bool)
	for _, p := range finalPlanets {
		if idSet[p.ID] {
			t.Errorf("Found duplicate planet ID: %d", p.ID)
		}
		idSet[p.ID] = true
	}
}
