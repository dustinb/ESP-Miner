package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"oldbute.com/axego/db"
	"oldbute.com/axego/service"
)

//go:embed dist
var frontend embed.FS

func main() {
	// Database connection setup
	db.InitDB()

	// Start network scanner
	foundAxe := make(chan db.Bitaxe)
	go func() {
		for bitaxe := range foundAxe {
			log.Printf("Found Axe: %s %s %s", bitaxe.Hostname, bitaxe.IP, bitaxe.MacAddr)
			db.Database.Save(&bitaxe)
		}
	}()
	go service.ScanNetwork(foundAxe)

	// Let the network scan finish
	time.Sleep(3 * time.Second)
	go service.TakeMeasurements()

	// Serve AxeOS frontend
	fs, _ := fs.Sub(frontend, "dist")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL)
		http.FileServer(http.FS(fs)).ServeHTTP(w, r)
	})

	http.HandleFunc("/axego/bitaxes", func(w http.ResponseWriter, r *http.Request) {
		var bitaxes []db.Bitaxe
		db.Database.Find(&bitaxes)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bitaxes)
	})

	http.HandleFunc("/axego/measurements", func(w http.ResponseWriter, r *http.Request) {
		var measurements []db.Measurement

		from := r.URL.Query().Get("from")
		query := db.Database.Where("mac_addr = ?", "E4:B0:63:86:72:C8")

		if from != "0" {
			unixTime, _ := strconv.ParseInt(from, 10, 64)
			fromTime := time.Unix(unixTime/1000, 0)

			query.Where("created_at > ?", fromTime).Order("created_at ASC")
			query.Find(&measurements)
		} else {
			query.Order("created_at desc").Limit(100).Find(&measurements)
			sort.Slice(measurements, func(i, j int) bool {
				return measurements[i].CreatedAt.Before(measurements[j].CreatedAt)
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(measurements)
	})

	// Create a route /api/* that acts as a proxy for /api requests to the current Bitaxe
	http.HandleFunc("/api/", func(w http.ResponseWriter, orig *http.Request) {
		client := &http.Client{}
		axeURL := "http://" + "192.168.1.226"

		var bodyBuffer bytes.Buffer
		if orig.Body != nil {
			io.Copy(&bodyBuffer, orig.Body)
		}

		req, _ := http.NewRequest(orig.Method, axeURL+orig.URL.Path, &bodyBuffer)
		req.Header = orig.Header

		response, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		// Copy the response headers to the proxy response
		for key, value := range response.Header {
			w.Header().Set(key, value[0])
		}

		w.WriteHeader(response.StatusCode)
		io.Copy(w, response.Body)
	})

	log.Println("Starting server on port 80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
