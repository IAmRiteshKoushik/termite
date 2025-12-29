package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rk/tentacloid/pkg"
)

type HackathonPayload struct {
	TeamName          string                `json:"team_name"`
	LeaderName        string                `json:"leader_name"`
	LeaderEmail       string                `json:"leader_email"`
	LeaderPhoneNumber string                `json:"leader_phone_number"`
	LeaderCollegeName string                `json:"leader_college_name"`
	ProblemStatement  string                `json:"problem_statement"`
	TeamMembers       []HackathonTeamMember `json:"team_members"`
}

type HackathonTeamMember struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	CollegeName string `json:"college_name"`
}

const (
	AIWebhookURL = "http://your-target-endpoint.com/ai-hackathon"
)

var aiClient = &http.Client{
	Timeout: time.Second * 10,
}

func DispatchHackathonPayload(payload HackathonPayload) error {
	pkg.Log.Info(fmt.Sprintf("Dispatching payload for team: %s", payload.TeamName))

	jsonData, err := json.Marshal(payload)
	if err != nil {
		pkg.Log.Error("Failed to marshal HackathonPayload", err)
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", AIWebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		pkg.Log.Error("Failed to create HTTP request", err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := aiClient.Do(req)
	if err != nil {
		pkg.Log.Error("Failed to dispatch payload", err)
		return fmt.Errorf("failed to dispatch request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		pkg.Log.Info(fmt.Sprintf("Successfully dispatched payload for team: %s, status: %s", payload.TeamName, resp.Status))
		return nil
	}

	pkg.Log.Warn(fmt.Sprintf("Failed to dispatch payload for team: %s, status: %s", payload.TeamName, resp.Status))
	return fmt.Errorf("request failed with status: %s", resp.Status)
}
