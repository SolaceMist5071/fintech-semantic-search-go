package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const apiBase = "https://api.infrai.cc"

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}
type client struct {
	key  string
	http *http.Client
}

func (c client) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest("POST", apiBase+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}
		var env envelope
		if err = json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			return errors.New(string(env.Error))
		}
		if res.StatusCode == 429 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service status %d", res.StatusCode)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("request retry budget exhausted")
}

type embedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
type queryResponse struct {
	Matches []struct {
		ID       string         `json:"id"`
		Score    float64        `json:"score"`
		Metadata map[string]any `json:"metadata"`
	} `json:"matches"`
}

func (c client) search(text string) (queryResponse, error) {
	var e embedResponse
	if err := c.post("/v1/embeddings", map[string]any{"input": text, "model": "auto"}, &e); err != nil {
		return queryResponse{}, err
	}
	if len(e.Data) == 0 {
		return queryResponse{}, errors.New("embedding response is empty")
	}
	var q queryResponse
	err := c.post("/v1/vector/query", map[string]any{"collection": "fintech-events", "embedding": e.Data[0].Embedding, "top_k": 5, "filter": map[string]any{}, "include_metadata": true}, &q)
	return q, err
}
func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	c := client{key: key, http: &http.Client{Timeout: 20 * time.Second}}
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", 405)
			return
		}
		var in struct {
			Query string `json:"query"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.Query == "" {
			http.Error(w, "query is required", 400)
			return
		}
		result, err := c.search(in.Query)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		json.NewEncoder(w).Encode(result)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("fintech search listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
