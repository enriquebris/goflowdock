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
		// compile the regex patterns
		for i := 0; i < len(cmd.Pattern); i++ {
			compiledPattern, err := regexp.Compile(cmd.Pattern[i])
			if err != nil {
				return err
			}

			// save the compiled regex pattern
			cmd.addCompiledRegex(compiledPattern)
		}
	}

	if cmd.PatternType == CMDTypeWord {
		// lowercase patterns
		for i := 0; i < len(cmd.Pattern); i++ {
			cmd.Pattern[i] = strings.ToLower(cmd.Pattern[i])
		}
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

		cmd, pattern, useHandler, cmdContent, err := st.matchCMDs(st.commands, entry.Content)
		if err != nil {
			if st.errorChan != nil {
				st.errorChan <- err
			}
		} else {
			if useHandler && cmd.Handler != nil {
				cmd.Handler(cmd, pattern, entry, cmdContent)
			}

			if !useHandler && cmd.ErrorHandler != nil {
				cmd.ErrorHandler(cmd, pattern, entry, cmdContent)
			}
		}
	}
}

// matchCMDs finds for the best CMD (command) match.
// Returns CMD, bool, err
// CMD		==> best match command
// string	==> pattern that matches
// bool		==> true (use CMD.handler), false (use CMD.ErrorHandler)
// string 	==> content
// err		==> error
func (st *StreamManager) matchCMDs(cmds []CMD, content string) (CMD, string, bool, string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return CMD{}, "", false, content, NewCMDError(CMDErrorTypeNoCommand, fmt.Sprintf("no commands for '%v'", content))
	}

	// split the content into words separated by blank spaces
	words := strings.Split(content, " ")

	var (
		cmd2ret      CMD
		patternMatch string
	)

	for _, cmd := range cmds {
		match := false
		extraContent := ""

		switch cmd.PatternType {
		// regex
		case CMDTypeRegex:
			for i := 0; i < len(cmd.compiledRegex); i++ {
				match = cmd.compiledRegex[i].MatchString(content)

				if match {
					// saved the pattern that matches
					patternMatch = cmd.Pattern[i]
					break
					// TODO ::: get the extracontent from the regex
				}
			}

		// first word
		case CMDTypeWord:
			for i := 0; i < len(cmd.Pattern); i++ {
				match = strings.ToLower(words[0]) == cmd.Pattern[i]
				if match {
					// saved the pattern that matches
					patternMatch = cmd.Pattern[i]
					// extra content to keep parsing
					extraContent = content[len(words[0]):]
					break
				}
			}
		}

		if match {
			cmd2ret = cmd

			extraContent = strings.TrimSpace(extraContent)

			// return CMD if no more content to parse
			// TODO :: check for required subCommands or Params
			if extraContent == "" {
				return cmd2ret, patternMatch, true, content, nil
			}

			checkForParams := true

			// check for subCommands
			if len(cmd.SubCommands) > 0 {
				// do not check for params
				checkForParams = false
				extraCmd2ret, patternMatch2, useHandler, content2, err2 := st.matchCMDs(cmd.SubCommands, extraContent)
				if err2 == nil {
					return extraCmd2ret, patternMatch2, useHandler, content2, nil
				}
			}

			// check for params
			if checkForParams {

			}

			// use errorHandler instead of handler
			return cmd2ret, patternMatch, false, extraContent, nil
		}

	}

	return CMD{}, "", false, content, NewCMDError(CMDErrorTypeNoCommand, fmt.Sprintf("no commands for '%v'", content))
}
