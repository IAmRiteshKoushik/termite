package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/IAmRiteshKoushik/termite/pkg"
)

type WoCPayload struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

var wocClient = &http.Client{
	Timeout: time.Second * 10,
}

// Convert incoming payload into JSON and dispatch to webhook URL
func DispatchWoCPayload(payload WoCPayload) error {
	pkg.Log.Info(fmt.Sprintf("Dispatching payload for email: %s", payload.Email))

	jsonData, err := json.Marshal(payload)
	if err != nil {
		pkg.Log.Error("Failed to marshal WoCPayload", err)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", pkg.AppConfig.WoC.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		pkg.Log.Error("Failed to create HTTP request", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Dispatch the request
	resp, err := wocClient.Do(req)
	if err != nil {
		pkg.Log.Error("Failed to dispatch payload", err)
		return fmt.Errorf("failed to dispatch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		pkg.Log.Info(fmt.Sprintf("Successfully dispatched payload for email: %s, status: %s", payload.Email, resp.Status))
		return nil
	}

	pkg.Log.Warn(fmt.Sprintf("Failed to dispatch payload for email: %s, status: %s", payload.Email, resp.Status))
	return fmt.Errorf("request failed with status: %s", resp.Status)
}
