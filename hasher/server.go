package hasher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Delay for calculating hashes
const (
	delaySeconds = 5 * time.Second
)

// ParseURLId
// Use regex to parse the id from "/hash/:id" properly
// and convert it to an int
func ParseUrlId(path string) (int, error) {
	re := regexp.MustCompile("^/\\w+/(\\d+)$")
	match := re.FindStringSubmatch(path)

	if len(match) < 2 {
		return 0, fmt.Errorf("id not found")
	}

	id, err := strconv.Atoi(match[1])

	if err != nil {
		return 0, err
	}

	return id, nil
}

// HashServer has internal http.Server type so it can ListenAndServe by itself
// Utilizes wait group for making go routines finish processing on shutdown
// Hasher for tracking generated hashes and stats
// Has a shutdown channel for, you know, shutting down
type HashServer struct {
	server    *http.Server
	waitGroup *sync.WaitGroup
	hasher    *Hasher
	shutdown  chan struct{}
}

// PostHashHandler
// POST /hash endpoint requiring a password paramter
// parse password from request, get the next id,
// immediately return it, generate hash and add
// id, hash, startTime to the Hasher after 5 seconds
// in the background
func (hs *HashServer) PostHashHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			timeNow := time.Now()

			r.ParseForm()

			// here password is a string slice
			password, ok := r.PostForm["password"]

			if !ok {
				http.Error(w, "password parameter is required", http.StatusBadRequest)
				return
			}

			id := hs.hasher.NextId()
			fmt.Fprintf(w, "%d", id)

			// waitGroup.Add must be called before the goroutine starts
			// to guarantee that Add happends before the goroutine that
			// calls waitGroup.Wait (when the http.func from GetShudwonHandler())
			// is invoked
			hs.waitGroup.Add(1)

			// generate hash from password parameter in the background with 5 sec delay
			go func() {
				defer hs.waitGroup.Done()
				log.Printf("id: %d, received password, waiting %s to generate hash\n", id, delaySeconds.String())
				time.Sleep(delaySeconds)
				hs.hasher.Add(id, password[0], timeNow)
			}()
		} else {
			http.Error(w, "POST method is required", http.StatusMethodNotAllowed)
		}
	}
}

// GetHashHandler
// GET /hash/:id enpoint wrapping the get call on Hasher with the id
// parsed from the url path and convert to int for hash lookup
func (hs *HashServer) GetHashHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			id, err := ParseUrlId(r.URL.Path)

			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			hash, err := hs.hasher.Get(id)

			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			fmt.Fprintf(w, "%s", hash)
		} else {
			http.Error(w, "GET method is required", http.StatusMethodNotAllowed)
		}
	}
}

// GetStatsHandler
// GET /stats endpoint wrapping hasher call to generate
// current stats on the fly and returning json
func (hs *HashServer) GetStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(hs.hasher.GenerateStats())
		} else {
			http.Error(w, "GET method is required", http.StatusMethodNotAllowed)
		}
	}
}

// GetShutdownHandler
// GET /shutdown endpoint. first, close all idle connections,
// then wait indefinitely for connections to return to idle
// and then shut down the server.
// second, wait for all go routines to finish processing
// lastly, close the shutdown channel so a thread running
// HashServer (usually a main thread) will terminate
func (hs *HashServer) GetShutdownHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			go func() {
				log.Println("Received Shutdown!, stop accepting new connections")
				err := hs.server.Shutdown(context.Background())

				if err != nil {
					log.Printf("Error Shutting Down HTTP Server: %v", err)
				}

				log.Println("Shutting down, waiting requests to finish...")
				hs.waitGroup.Wait()
				close(hs.shutdown)
			}()
		} else {
			http.Error(w, "GET method is required", http.StatusMethodNotAllowed)
		}
	}
}

// ListenAndServe
// delegate ListenAndServe down to our own http.Server
func (hs *HashServer) ListenAndServe() error {
	log.Println("Servin' them hashes HOT!")
	return hs.server.ListenAndServe()
}

// NewHashServer
// HashServer's state is halfway built after initializing with wait group,
// newHasher and shutdown channel. Second half is building a newServeMux
// and setting handlers from newHashServer's current state of handler functions
// that wrap http.HandlerFunc.
// Store newMuxServer in a new http.Server as its handler along with constructed
// addressPort string (from user input) which allows us to call ListenAndServe
// on it.
// Return our new HashServer with our receive only  shutdown channel to the caller (usually main)
func NewHashServer(address string, port string) (*HashServer, <-chan struct{}) {
	addressPort := address + ":" + port
	shutdown := make(chan struct{})

	newHashServer := &HashServer{
		waitGroup: &sync.WaitGroup{},
		hasher:    NewHasher(),
		shutdown:  shutdown,
	}

	newMux := http.NewServeMux()
	newMux.Handle("/hash", newHashServer.PostHashHandler())
	newMux.Handle("/hash/", newHashServer.GetHashHandler())
	newMux.Handle("/stats", newHashServer.GetStatsHandler())
	newMux.Handle("/shutdown", newHashServer.GetShutdownHandler())

	newHashServer.server = &http.Server{Addr: addressPort, Handler: newMux}

	return newHashServer, shutdown
}
