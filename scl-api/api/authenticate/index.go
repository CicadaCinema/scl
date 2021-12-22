package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type AuthRequest struct {
	ClientCode string `json:"client_code"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

type StravaAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    int    `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var err error
	var authRequest AuthRequest
	var stravaAuthResponse StravaAuthResponse

	startTime := time.Now()

	// set essential headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// ensure we are receiving a post request
	fmt.Println("DEBUG: incoming", r.Method)
	if r.Method == "OPTIONS" {
		// this request just needed the headers above
		return
	} else if r.Method != "POST" {
		http.Error(w, "Method is invalid", http.StatusMethodNotAllowed)
		return
	}

	// process request body, populate authRequest
	if err := json.NewDecoder(r.Body).Decode(&authRequest); err != nil {
		http.Error(w, "Unable to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// get token from Strava, populate stravaAuthResponse
	resp, err := http.Post(
		fmt.Sprintf(
			"https://www.strava.com/api/v3/oauth/token?client_id=%s&client_secret=%s&code=%s&grant_type=%s",
			os.Getenv("STRAVA_CLIENT_ID"),
			os.Getenv("STRAVA_CLIENT_SECRET"),
			authRequest.ClientCode,
			"authorization_code",
		),
		"text/plain",
		bytes.NewBuffer([]byte("")),
	)

	if err != nil {
		http.Error(w, "Unable to authenticate with Strava: "+err.Error(), http.StatusInternalServerError)
		return
	} else if stravaResponseBody, err := ioutil.ReadAll(resp.Body); err != nil {
		http.Error(w, "Unable to read Strava authentication response: "+err.Error(), http.StatusInternalServerError)
		return
	} else if err = json.Unmarshal(stravaResponseBody, &stravaAuthResponse); err != nil {
		println("DEBUG: " + string(stravaResponseBody))
		http.Error(w, "Unable to unmarshal Strava authentication response: "+err.Error(), http.StatusInternalServerError)
		return
	} else if stravaAuthResponse.AccessToken == "" {
		http.Error(w, "Unable to get Strava auth code.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	json.NewEncoder(w).Encode(AuthResponse{AccessToken: stravaAuthResponse.AccessToken})

	fmt.Println("DEBUG: this successful request took", time.Since(startTime).Milliseconds(), "ms")
}
