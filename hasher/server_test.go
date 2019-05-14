package hasher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// test for valid and invalid url paths in order to parse id correctly
func TestParseUrlId(t *testing.T) {
	UrlCases := []struct {
		url string
		id  int
		err bool
	}{
		{"/hash/1", 1, false},
		{"/hash/123456789", 123456789, false},
		{"/hash/-1", 0, true},
		{"/hash/1/2", 0, true},
		{"/123a", 0, true},
		{"/blahblah/", 0, true},
		{"blahblah/", 0, true},
	}
	for _, c := range UrlCases {
		id, err := ParseUrlId(c.url)
		if id != c.id || (err == nil && c.err == true) || (err != nil && c.err == false) {
			t.Errorf("UrlID(%s): expected(id, err): %d, %t, got: %d, %t", c.url, c.id, c.err, id, err)
		}
	}
}

// send wrong http request to each endpoint to verify correct status code
func TestStatusMethodNotAllowed(t *testing.T) {
	hs, _ := NewHashServer("localhost", "8080")

	// send get to post /hash endpoint
	req := httptest.NewRequest("GET", "/hash", nil)
	w := httptest.NewRecorder()

	postHashHandler := hs.PostHashHandler()
	postHashHandler(w, req)

	if status := w.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusMethodNotAllowed, status)
	}

	postData := url.Values{}
	postData.Add("somePostParam", "someValue")
	encodedPostDataReader := strings.NewReader(postData.Encode())

	// send post to get /hash/:id endpoint
	req = httptest.NewRequest("POST", "/hash/1", encodedPostDataReader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	getHashHandler := hs.GetHashHandler()
	getHashHandler(w, req)

	if status := w.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusMethodNotAllowed, status)
	}

	// send post to get /stats endpoint
	req = httptest.NewRequest("POST", "/stats", encodedPostDataReader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	getStatsHandler := hs.GetStatsHandler()
	getStatsHandler(w, req)

	if status := w.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusMethodNotAllowed, status)
	}

	// send post to get /shutdown endpoint
	req = httptest.NewRequest("POST", "/shtudown", encodedPostDataReader)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	getShutdownHandler := hs.GetShutdownHandler()
	getShutdownHandler(w, req)

	if status := w.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusMethodNotAllowed, status)
	}
}

func TestHashServer(t *testing.T) {
	hs, shutdown := NewHashServer("localhost", "8080")
	zero := int64(0)

	// test unpopulated stats and valid json is returned for HashStats reconstruction
	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	getStatsHandler := hs.GetStatsHandler()
	getStatsHandler(w, req)

	var validZeroHashStats HashStats
	err := json.Unmarshal([]byte(w.Body.String()), &validZeroHashStats)

	if err != nil {
		t.Errorf("expected: valid json to be HashStats type got: %v", err.Error())
	}
	if total := validZeroHashStats.Total; total != zero {
		t.Errorf("expected: %d, got: %d", zero, total)
	}
	if average := validZeroHashStats.Average; average != zero {
		t.Errorf("expected: %d, got: %d", zero, average)
	}

	// test posting password to /hash endpoint and recieving the correct id
	postHashData := url.Values{}
	postHashData.Add("password", "angryMonkey")

	req = httptest.NewRequest("POST", "/hash", strings.NewReader(postHashData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	postHashHandler := hs.PostHashHandler()
	postHashHandler(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusOK, status)
	}

	expected := "1"
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: expected: %v, got: %v", expected, w.Body.String())
	}

	// query right away for hash from the id we received
	// which should not exist because of 5 second delay
	req = httptest.NewRequest("GET", "/hash", nil)
	w = httptest.NewRecorder()

	getHashHandler := hs.GetHashHandler()
	getHashHandler(w, req)

	if status := w.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusNotFound, status)
	}

	expected = "id not found\n"
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: expected: %v, got: %v", expected, w.Body.String())
	}

	// wait at least 5 seconds for password to be hashedi and receive correct hash
	time.Sleep(6 * time.Second)
	req = httptest.NewRequest("GET", "/hash/1", nil)
	w = httptest.NewRecorder()

	getHashHandler = hs.GetHashHandler()
	getHashHandler(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: expected: %v, got: %v", http.StatusOK, status)
	}

	expected = "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: expected: %v, got: %v", expected, w.Body.String())
	}

	// test populated stats and valid json is returned for HashStats reconstruction
	req = httptest.NewRequest("GET", "/stats", nil)
	w = httptest.NewRecorder()

	getStatsHandler = hs.GetStatsHandler()
	getStatsHandler(w, req)

	var validHashStats HashStats
	err = json.Unmarshal([]byte(w.Body.String()), &validHashStats)

	if err != nil {
		t.Errorf("expected: valid json to be HashStats type got: %v", err.Error())
	}
	if total := validHashStats.Total; total != 1 {
		t.Errorf("expected: %v got: %v", 1, total)
	}
	// couldnt figure out to test for exact stats here but the test for Hasher.GenerateStats() should suffice
	if average := validHashStats.Average; !(average >= int64(5000) && average <= int64(5006)) {
		t.Errorf("expected: between %v and %v, got: %v", 5000, 5006, average)
	}

	// test shutdown and see if shutdown channel is closed
	req = httptest.NewRequest("GET", "/shutdown", nil)
	w = httptest.NewRecorder()

	getShutdownHandler := hs.GetShutdownHandler()
	getShutdownHandler(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	// wait for shutdown channel to be closed from /shutdown endpoint
	time.Sleep(1 * time.Second)

	select {
	case <-shutdown:
	default:
		t.Errorf("expected: shutdown channel to close")
	}
}
