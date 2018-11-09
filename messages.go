package goflowdock

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type MessageData struct {
	Flow             string   `json:"flow"`
	Event            string   `json:"event"`
	Content          string   `json:"content"`
	Tags             []string `json:"tags,omitempty"`
	ExternalUserName string   `json:"external_user_name"`
	ThreadID         string   `json:"thread_id"`
}

type MessageManager struct {
	token        string
	organization string
}

func NewMessageManager(token string, organization string) *MessageManager {
	ret := &MessageManager{}
	ret.initialize(token, organization)

	return ret
}

func (st *MessageManager) initialize(token string, organization string) {
	st.token = b64.StdEncoding.EncodeToString([]byte(token))
	st.organization = organization
}

func (st *MessageManager) SendMessage(message MessageData) (*http.Response, error) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	url := URLMessages

	request, err := http.NewRequest("POST", url, bytes.NewReader(jsonMessage))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Basic %v", st.token))

	client := http.Client{}
	return client.Do(request)
}
