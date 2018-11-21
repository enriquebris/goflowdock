package goflowdock

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// A Flowdock entry structure
// Each streamed response will be transformed into an event
type EntryOld struct {
	Event       string    `json:"event"`
	Tags        []string  `json:"tags"`
	Uuid        string    `json:"uuid"`
	Persist     bool      `json:"persist"`
	Id          int       `json:"id"`
	Flow        string    `json:"flow"`
	Content     string    `json:"content"`
	Sent        int       `json:"sent"`
	App         string    `json:"app"`
	CreatedAt   time.Time `json:"created_at"`
	Attachments []string  `json:"attachments"`
	User        int       `json:"user"`
	ThreadId    string    `json:"thread_id,omitempty"`
}

type Entry struct {
	ThreadID string        `json:"thread_id,omitempty"`
	Event    string        `json:"event"`
	Tags     []interface{} `json:"tags,omitempty"`
	UUID     string        `json:"uuid"`
	Thread   struct {
		Body             string        `json:"body"`
		ExternalURL      interface{}   `json:"external_url"`
		ExternalComments int           `json:"external_comments"`
		Activities       int           `json:"activities"`
		ID               string        `json:"id"`
		Flow             string        `json:"flow"`
		Status           interface{}   `json:"status"`
		Fields           []interface{} `json:"fields"`
		Actions          []interface{} `json:"actions"`
		CreatedAt        int64         `json:"created_at"`
		Title            string        `json:"title"`
		UpdatedAt        time.Time     `json:"updated_at"`
		InitialMessage   int           `json:"initial_message"`
		InternalComments int           `json:"internal_comments"`
	} `json:"thread,omitempty"`
	ID             int           `json:"id"`
	Flow           string        `json:"flow"`
	Content        string        `json:"content"`
	Sent           int64         `json:"sent"`
	App            string        `json:"app"`
	CreatedAt      time.Time     `json:"created_at"`
	Attachments    []interface{} `json:"attachments"`
	EmojiReactions struct {
	} `json:"emojiReactions,omitempty"`
	User string `json:"user"`
}

type StreamManager struct {
	authToken string
	errorChan chan error
}

func NewStreamManager(authToken string, errorChan chan error) *StreamManager {
	ret := &StreamManager{}
	ret.initialize(authToken, errorChan)

	return ret
}

func (st *StreamManager) initialize(authToken string, errorChan chan error) {
	st.authToken = b64.StdEncoding.EncodeToString([]byte(authToken))
	st.errorChan = errorChan
}

func (st *StreamManager) Listen(streamURL string, inputHandler func(entry Entry)) error {
	request, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Basic %v", st.authToken))

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(response.Body)

	var (
		errStream error = nil
		line      []byte
	)
	for {
		line, errStream = reader.ReadBytes('\n')
		if errStream != nil {
			return err
		}

		var entry Entry
		err = json.Unmarshal(line, &entry)

		if err == nil {
			// call input handler
			inputHandler(entry)
		} else {
			if st.errorChan != nil {
				st.errorChan <- err
			}
		}
	}

	return nil
}
