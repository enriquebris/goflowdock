package goflowdock

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// A Flowdock entry structure
// Each streamed response will be transformed into an event
type Entry struct {
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

type StreamManager struct {
	authToken string
	commands  []CMD
	errorChan chan error
}

func NewStreamManager(authToken string, errorChan chan error) *StreamManager {
	ret := &StreamManager{}
	ret.initialize(authToken, errorChan)

	return ret
}

func (st *StreamManager) initialize(authToken string, errorChan chan error) {
	st.authToken = b64.StdEncoding.EncodeToString([]byte(authToken))
	st.commands = make([]CMD, 0)
	st.errorChan = errorChan
}

func (st *StreamManager) AddCommand(cmd CMD) error {
	// compile && save the regex pattern
	if cmd.PatternType == CMDTypeRegex {
		// compile the regex pattern
		compiledPattern, err := regexp.Compile(cmd.Pattern)
		if err != nil {
			return err
		}

		// save the compiled regex pattern
		cmd.compiledRegex = compiledPattern
	}

	if cmd.PatternType == CMDTypeWord {
		cmd.Pattern = strings.ToLower(cmd.Pattern)
	}

	// save the command
	st.commands = append(st.commands, cmd)

	return nil
}

func (st *StreamManager) Listen(streamURL string) error {
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

		cmd, useHandler, cmdContent, err := st.matchCMDs(st.commands, entry.Content)
		if err != nil {
			if st.errorChan != nil {
				st.errorChan <- err
			}
		} else {
			if useHandler && cmd.Handler != nil {
				cmd.Handler(cmd, entry, cmdContent)
			}

			if !useHandler && cmd.ErrorHandler != nil {
				cmd.ErrorHandler(cmd, entry, cmdContent)
			}
		}
	}
}

// matchCMDs finds for the best CMD (command) match.
// Returns CMD, bool, err
// CMD		==> best match command
// bool		==> true (use CMD.handler), false (use CMD.ErrorHandler)
// string 	==> content
// err		==> error
func (st *StreamManager) matchCMDs(cmds []CMD, content string) (CMD, bool, string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return CMD{}, false, content, NewCMDError(CMDErrorTypeNoCommand, fmt.Sprintf("no commands for '%v'", content))
	}

	// split the content into words separated by blank spaces
	words := strings.Split(content, " ")

	var cmd2ret CMD

	for _, cmd := range cmds {
		match := false
		extraContent := ""

		switch cmd.PatternType {
		// regex
		case CMDTypeRegex:
			match = cmd.compiledRegex.MatchString(content)
			// TODO ::: get the extracontent from the regex

		// first word
		case CMDTypeWord:
			match = strings.ToLower(words[0]) == cmd.Pattern
			if match {
				extraContent = content[len(words[0]):]
			}
		}

		if match {
			cmd2ret = cmd

			extraContent = strings.TrimSpace(extraContent)

			// return CMD if no more content to parse
			// TODO :: check for required subCommands or Params
			if extraContent == "" {
				return cmd2ret, true, content, nil
			}

			checkForParams := true

			// check for subCommands
			if len(cmd.SubCommands) > 0 {
				// do not check for params
				checkForParams = false
				extraCmd2ret, useHandler, content2, err2 := st.matchCMDs(cmd.SubCommands, extraContent)
				if err2 == nil {
					return extraCmd2ret, useHandler, content2, nil
				}
			}

			// check for params
			if checkForParams {

			}

			// use errorHandler instead of handler
			return cmd2ret, false, extraContent, nil
		}

	}

	return CMD{}, false, content, NewCMDError(CMDErrorTypeNoCommand, fmt.Sprintf("no commands for '%v'", content))
}
