package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type PubSubMessage struct {
	Message struct {
		Data       []byte            `json:"data"`
		Attributes map[string]string `json:"attributes"`
	} `json:"message"`
}

type AlertPayload struct {
	Incident struct {
		AlertingPolicyName string `json:"policy_name"`
		Summary            string `json:"summary"`
		Url                string `json:"url"`
		ConditionName      string `json:"condition_name"`
		Metric             struct {
			Labels map[string]string `json:"labels"`
		} `json:"metric"`
		Resource struct {
			Labels map[string]string `json:"labels"`
			Type   string
		} `json:"resource"`
	} `json:"incident"`
}

type TeamsMessage struct {
	Type       string            `json:"type"`
	Context    string            `json:"context"`
	ThemeColor string            `json:"themeColor"`
	Summary    string            `json:"summary"`
	Sections   []Section         `json:"sections"`
	Actions    []PotentialAction `json:"potentialAction"`
}

type Section struct {
	ActivityTitle    string `json:"activityTitle"`
	ActivitySubtitle string `json:"activitySubtitle"`
	Facts            []Fact `json:"facts"`
}

type Fact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PotentialAction struct {
	Type    string   `json:"@type"`
	Name    string   `json:"name"`
	Targets []Target `json:"targets"`
}

type Target struct {
	Os  string `json:"os"`
	Uri string `json:"uri"`
}

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	webhookUrl := os.Getenv("WEBHOOK_URL")
	if len(webhookUrl) == 0 {
		http.Error(w, "no webhook url configured", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read body: %v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var message PubSubMessage
	if err := json.Unmarshal(body, &message); err != nil {
		log.Printf("failed to unmarshal pubsub message: %v", err)
		http.Error(w, "failed to unmarshal pubsub message", http.StatusBadRequest)
		return
	}

	var alertPayload AlertPayload
	if err := json.Unmarshal(message.Message.Data, &alertPayload); err != nil {
		log.Printf("failed to unmarshal incident payload: %v", err)
		http.Error(w, "failed to unmarshal incident payload", http.StatusBadRequest)
		return
	}

	log.Printf("received alert %+v", alertPayload)

	teamsMessage := generateMessageFromPayload(alertPayload)
	err = sendMessage(webhookUrl, teamsMessage)
	if err != nil {
		log.Printf("failed to send message to webhook: %v", err)
		http.Error(w, "failed to send message to webhook", http.StatusInternalServerError)
		return
	}
}

func generateMessageFromPayload(payload AlertPayload) TeamsMessage {
	facts := []Fact{}

	for labelKey, labelValue := range payload.Incident.Resource.Labels {
		facts = append(facts, Fact{Name: labelKey, Value: labelValue})
	}

	return TeamsMessage{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: "B30000",
		Summary:    payload.Incident.AlertingPolicyName,
		Sections: []Section{
			{
				ActivityTitle:    payload.Incident.ConditionName,
				ActivitySubtitle: payload.Incident.Summary,
				Facts:            facts,
			},
		},
		Actions: []PotentialAction{
			{
				Type: "OpenUri",
				Name: "View Incident",
				Targets: []Target{
					{
						Os:  "default",
						Uri: payload.Incident.Url,
					},
				},
			},
		},
	}
}

func sendMessage(url string, message TeamsMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = http.Post(url, "application/json", bytes.NewReader(body))
	return err
}
