package goflowdock

import "regexp"

const (
	CMDTypeRegex = "regex"
	CMDTypeword  = "word"
)

// CMDHandler is the function handler
type CMDHandler func(cmd CMD, entry Entry, cmdContent string)

// CMD command
type CMD struct {
	PatternType   string
	Pattern       string
	compiledRegex *regexp.Regexp
	Handler       CMDHandler
	ErrorHandler  CMDHandler
	SubCommands   []CMD
}

// ***********************************************************************************************
// **  CMDError  *********************************************************************************
// ***********************************************************************************************

type CMDError struct {
	errorMessage string
	errorType    string
}

const (
	CMDErrorTypeNoCommand = "noCommand"
)

func NewCMDError(errorType string, errorMessage string) *CMDError {
	return &CMDError{
		errorType:    errorType,
		errorMessage: errorMessage,
	}
}

func (st *CMDError) Error() string {
	return st.errorMessage
}

func (st *CMDError) GetType() string {
	return st.errorType
}
