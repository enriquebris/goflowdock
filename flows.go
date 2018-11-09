package goflowdock

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Flows []struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ParameterizedName string    `json:"parameterized_name"`
	Email             string    `json:"email"`
	Description       string    `json:"description"`
	URL               string    `json:"url"`
	WebURL            string    `json:"web_url"`
	AccessMode        string    `json:"access_mode"`
	FlowAdmin         bool      `json:"flow_admin"`
	APIToken          string    `json:"api_token"`
	Open              bool      `json:"open"`
	Joined            bool      `json:"joined"`
	LastMessageAt     time.Time `json:"last_message_at"`
	LastMessageID     int       `json:"last_message_id"`
	TeamNotifications bool      `json:"team_notifications"`
	Organization      struct {
		Active            bool   `json:"active"`
		BillingMethod     string `json:"billing_method"`
		ID                int    `json:"id"`
		Name              string `json:"name"`
		ParameterizedName string `json:"parameterized_name"`
		URL               string `json:"url"`
		UserCount         int    `json:"user_count"`
		UserLimit         int    `json:"user_limit"`
		FlowAdmins        bool   `json:"flow_admins"`
	} `json:"organization"`
	FlowGroups []interface{} `json:"flow_groups"`
}

type FlowManager struct {
	token string
}

func NewFlowManager(token string) *FlowManager {
	ret := &FlowManager{}
	ret.initialize(token)

	return ret
}

func (st *FlowManager) initialize(token string) {
	st.token = b64.StdEncoding.EncodeToString([]byte(token))
}

func (st *FlowManager) SetToken(token string) {
	st.token = b64.StdEncoding.EncodeToString([]byte(token))
}

// GetFlows read and returns all visible flows (based on the API token).
func (st *FlowManager) GetFlows() (Flows, error) {
	request, err := http.NewRequest("GET", URLFlows, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Basic %v", st.token))

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	s := buf.String()

	var flows Flows
	err = json.Unmarshal([]byte(s), &flows)
	if err != nil {
		return nil, err
	}

	return flows, nil
}
