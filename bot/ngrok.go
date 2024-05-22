package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	maxAttempts      = 10
	sleepIntervalSec = 5
)

type NgrokTunnel struct {
	PublicURL string `json:"public_url"`
}

type NgrokTunnelsResponse struct {
	Tunnels []NgrokTunnel `json:"tunnels"`
}

func getNgrokPublicURL(ngrokAPIURL string) (string, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := http.Get(ngrokAPIURL)
		if err != nil {
			log.Printf("Attempt %d/%d: Error fetching Ngrok API: %v", attempt, maxAttempts, err)
			time.Sleep(sleepIntervalSec * time.Second)
			continue
		}

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(resp.Body)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Attempt %d/%d: Error reading response body: %v", attempt, maxAttempts, err)
			time.Sleep(sleepIntervalSec * time.Second)
			continue
		}

		var ngrokResponse NgrokTunnelsResponse
		err = json.Unmarshal(body, &ngrokResponse)
		if err != nil {
			log.Printf("Attempt %d/%d: Error unmarshalling JSON: %v", attempt, maxAttempts, err)
			time.Sleep(sleepIntervalSec * time.Second)
			continue
		}

		if len(ngrokResponse.Tunnels) > 0 {
			return ngrokResponse.Tunnels[0].PublicURL, nil
		}

		log.Printf("Attempt %d/%d: No tunnels found, retrying in %d seconds...", attempt, maxAttempts, sleepIntervalSec)
		time.Sleep(sleepIntervalSec * time.Second)
	}

	return "", fmt.Errorf("failed to get Ngrok public URL after %d attempts", maxAttempts)
}
