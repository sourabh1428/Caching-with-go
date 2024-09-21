package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type InMemoryStore struct {
	data    map[string]string
	maxSize int
	mu      sync.Mutex // To handle concurrent access
}

func NewInMemoryStore(maxSize int) *InMemoryStore {
	return &InMemoryStore{
		data:    make(map[string]string),
		maxSize: maxSize,
	}
}

func (store *InMemoryStore) Add(key, value string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.data) >= store.maxSize {
		return false
	}
	store.data[key] = value
	return true
}

func (store *InMemoryStore) Get(key string) (string, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()

	value, exists := store.data[key]
	return value, exists
}

func (store *InMemoryStore) Delete(key string) bool {
	store.mu.Lock()
	defer store.mu.Unlock()

	_, exists := store.data[key]
	if exists {
		delete(store.data, key)
	}
	return exists
}

func main() {
	maxSize := 100
	store := NewInMemoryStore(maxSize)

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			fmt.Println("error in post request")
			return
		}

		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if store.Add(req.Key, req.Value) {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "Added: %s -> %s\n", req.Key, req.Value)
		} else {
			http.Error(w, "Store is full", http.StatusConflict)
		}
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		value, exists := store.Get(key)
		if exists {
			fmt.Fprintf(w, "Retrieved: %s -> %s\n", key, value)
		} else {
			http.Error(w, "Key not found", http.StatusNotFound)
		}
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if store.Delete(key) {
			fmt.Fprintf(w, "Deleted key: %s\n", key)
		} else {
			http.Error(w, "Key not found", http.StatusNotFound)
		}
	})

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
