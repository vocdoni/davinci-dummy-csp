package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/davinci-node/crypto/csp"
	"github.com/vocdoni/davinci-node/types"
	"github.com/vocdoni/davinci-node/util"
)

func main() {
	// Load the seed from an environment variable
	strSeed := os.Getenv("CSP_SEED")
	if strSeed == "" {
		log.Fatal("CSP_SEED environment variable is required")
	}
	strHTTPPort := os.Getenv("CSP_PORT")
	if strHTTPPort == "" {
		strHTTPPort = "8080" // Default port
	}

	// Create a new CSP with the correct origin and seed provided
	dummyCSP, err := csp.New(types.CensusOriginCSPEdDSABLS12377, []byte(strSeed))
	if err != nil {
		log.Fatalf("Error creating CSP: %v", err)
	}

	// Calculate the census root
	censusRoot := dummyCSP.CensusRoot()
	log.Printf("CSP initialized, census root: %x", censusRoot.Root)

	// Endpoint to return the census root
	http.HandleFunc("/root", func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(censusRoot)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})

	// Endpoint to get a census proof for a given process ID and address
	http.HandleFunc("/proof", func(w http.ResponseWriter, r *http.Request) {
		strPID := r.URL.Query().Get("pid")
		strAddr := r.URL.Query().Get("addr")
		if strPID == "" || strAddr == "" {
			http.Error(w, "Missing pid or addr parameter", http.StatusBadRequest)
			return
		}
		strPID = util.TrimHex(strPID)
		strAddr = util.TrimHex(strAddr)

		bPID, err := hex.DecodeString(strPID)
		if err != nil {
			http.Error(w, "Invalid pid parameter", http.StatusBadRequest)
			return
		}
		pid := new(types.ProcessID).SetBytes(bPID)
		if !pid.IsValid() {
			http.Error(w, "Invalid process ID", http.StatusBadRequest)
			return
		}

		addr := common.HexToAddress(strAddr)
		proof, err := dummyCSP.GenerateProof(pid, addr)
		if err != nil {
			http.Error(w, "Failed to generate proof: "+err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(proof)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})

	// Start the HTTP server
	log.Println("Starting server on :" + strHTTPPort)
	if err := http.ListenAndServe("0.0.0.0:"+strHTTPPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
