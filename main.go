package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Planet represents a planet in our solar system
type Planet struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	planets []Planet
	mutex   = &sync.Mutex{}
	nextID  = 1
)

func getPlanets(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planets)
}

func createPlanet(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	var planet Planet
	if err := json.NewDecoder(r.Body).Decode(&planet); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	planet.ID = nextID
	nextID++
	planets = append(planets, planet)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(planet)
}

func main() {
	planets = append(planets, Planet{ID: nextID, Name: "Mercury"})
	nextID++
	planets = append(planets, Planet{ID: nextID, Name: "Venus"})
	nextID++
	planets = append(planets, Planet{ID: nextID, Name: "Earth"})
	nextID++

	http.HandleFunc("/planets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPlanets(w, r)
		case http.MethodPost:
			createPlanet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/heavy", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Starting heavy task")
		time.Sleep(5 * time.Second)
		fmt.Fprintf(w, "Heavy task complete!")
		log.Println("Finished heavy task")
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
