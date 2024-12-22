package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

// Struct for handling the request body (domain)
type RequestBody struct {
	Domain string `json:"domain"`
}

// Struct for holding the subdomain and ports info
type Subdomain struct {
	Name  string `json:"name"`
	Ports []int  `json:"ports"`
}

// Struct for the API response
type Response struct {
	Subdomains []Subdomain `json:"subdomains"`
}

// Run subfinder to find subdomains for a given domain
func runSubfinder(domain string) ([]string, error) {
	log.Printf("Running subfinder for domain: %s", domain)
	cmd := exec.Command("subfinder", "-d", domain)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("subfinder error: %v", err)
	}

	var subdomains []string
	output := string(cmdOutput)
	log.Printf("Subfinder output: %s", output)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			subdomains = append(subdomains, line)
		}
	}
	return subdomains, nil
}

// Run nmap to get open ports for a given subdomain
func runNmap(subdomain string) ([]int, error) {
	log.Printf("Running nmap for subdomain: %s", subdomain)
	cmd := exec.Command("nmap", subdomain)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nmap error: %v", err)
	}

	var ports []int
	output := string(cmdOutput)
	log.Printf("Nmap output: %s", output)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "open") {
			var port int
			fmt.Sscanf(line, "%d/tcp", &port)
			ports = append(ports, port)
		}
	}
	return ports, nil
}

// ScanHandler handles the /scan endpoint to process domain and return subdomains and ports
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers to allow requests from specific origin (React frontend)
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle the preflight request (OPTIONS)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only handle POST requests for scanning
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var body RequestBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Println("Received domain for scanning:", body.Domain)

	// Run subfinder to get subdomains
	subdomains, err := runSubfinder(body.Domain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error running Subfinder: %v", err), http.StatusInternalServerError)
		return
	}

	// For each subdomain, run nmap to get open ports
	var subdomainData []Subdomain
	for _, subdomain := range subdomains {
		ports, err := runNmap(subdomain)
		if err != nil {
			log.Printf("Error running Nmap for %s: %v", subdomain, err)
			continue
		}
		subdomainData = append(subdomainData, Subdomain{Name: subdomain, Ports: ports})
	}

	// Send the response with subdomains and ports
	response := Response{
		Subdomains: subdomainData,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()

	// Handle the /scan endpoint
	r.HandleFunc("/scan", ScanHandler).Methods("POST", "OPTIONS")

	// Enable CORS handling using the gorilla/handlers package
	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
