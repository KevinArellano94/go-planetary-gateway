// Filename: main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Planet defines the structure for planetary data.
// This is a well-defined and extensible model, a best practice for API design.
type Planet struct {
	Name     string  `json:"name"`
	Diameter float64 `json:"diameter_km"`
	Mass     float64 `json:"mass_10e24_kg"`
}

// mockPlanetDatabase simulates a persistent data source like a SQL database or another microservice.
// Using a map allows for quick and easy lookups in this example.
var mockPlanetDatabase = map[string]Planet{
	"Mercury": {Name: "Mercury", Diameter: 4879, Mass: 0.330},
	"Venus":   {Name: "Venus", Diameter: 12104, Mass: 4.87},
	"Earth":   {Name: "Earth", Diameter: 12756, Mass: 5.97},
	"Mars":    {Name: "Mars", Diameter: 6792, Mass: 0.642},
	"Jupiter": {Name: "Jupiter", Diameter: 142984, Mass: 1898},
	"Saturn":  {Name: "Saturn", Diameter: 120536, Mass: 568},
	"Uranus":  {Name: "Uranus", Diameter: 51118, Mass: 86.8},
	"Neptune": {Name: "Neptune", Diameter: 49528, Mass: 102},
}

// fetchPlanetData simulates a slow I/O operation (e.g., a network or database call).
// The intentional delay makes the benefits of concurrency clearly observable.
func fetchPlanetData(name string) (Planet, bool) {
	// Simulate network latency to a downstream service.
	time.Sleep(800 * time.Millisecond)
	// Normalize input for a case-insensitive lookup.
	planet, found := mockPlanetDatabase[strings.Title(strings.ToLower(name))]
	return planet, found
}

// planetsHandler is the core of our API. It handles incoming requests to the /planets endpoint.
// It showcases a robust concurrent pattern to fetch data efficiently, perfect for 10x engineering.
func planetsHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse and validate the incoming request.
	namesQuery := r.URL.Query().Get("names")
	if namesQuery == "" {
		http.Error(w, "Query parameter 'names' is required (e.g., /planets?names=Earth,Mars).", http.StatusBadRequest)
		return
	}
	planetNames := strings.Split(namesQuery, ",")

	// Step 2: Set up concurrency primitives.
	// A WaitGroup is used to wait for all concurrent operations to complete.
	var wg sync.WaitGroup
	// A buffered channel safely collects the results from multiple goroutines.
	resultsChan := make(chan Planet, len(planetNames))

	// Step 3: Fan-out - launch a goroutine for each planet to fetch data concurrently.
	for _, name := range planetNames {
		// Ensure we have a valid name to process.
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" {
			continue
		}

		wg.Add(1) // Increment the WaitGroup counter before starting the goroutine.
		go func(planetName string) {
			defer wg.Done() // Decrement the counter when the goroutine completes.
			if planet, found := fetchPlanetData(planetName); found {
				resultsChan <- planet // Send the successful result to the channel.
			}
		}(trimmedName) // Pass the name as an argument to avoid a common closure bug.
	}

	// Step 4: Create a monitor goroutine to close the channel once all workers are done.
	// This is a critical step to signal the results collection loop that it can safely terminate.
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Step 5: Fan-in - collect all results from the channel until it's closed.
	var planets []Planet
	for planet := range resultsChan {
		planets = append(planets, planet)
	}

	// Step 6: Respond to the client with the collected data.
	// Setting the correct Content-Type header is a fundamental best practice.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planets)
}

func main() {
	// Register our handler for the /planets API endpoint.
	http.HandleFunc("/planets", planetsHandler)

	log.Println("🚀 Starting Planetary Gateway on http://localhost:8080")
	log.Println("Usage: http://localhost:8080/planets?names=Earth,Mars,Jupiter")

	// Start the HTTP server. Proper error handling for startup is essential for robust services.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("❌ Could not start server: %s\n", err)
	}
}
