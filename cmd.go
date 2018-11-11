package goflowdock

import (
	"regexp"
	"strconv"
)

const (
	CMDTypeRegex = "regex"
	CMDTypeWord  = "word"

	CMDHandlerTypeDefault       = "handler.default"
	CMDHandlerTypeError         = "handler.error"
	CMDHandlerTypeParams        = "handler.params"
	CMDHandlerTypeParamsMissing = "handler.params.missing"
	CMDHandlerTypeParamsExtra   = "handler.params.extra"
)

// CMDHandler is the function handler
//
// Parameters:
//
// cmd CMD				==> CMD that matched
// pattern string		==> exact CMD pattern that matched
// entry Entry			==> Related Flowdock entry
// cmdContent string	==> Content to parse (original content - cmd)
// handlerType string	==> Indicates which handler type will be used and gives some extra data (parameters: ok, missing, extra)
type CMDHandler func(cmd CMD, pattern string, entry Entry, cmdContent string, handlerType string)

// ***********************************************************************************************
// **  CMD  **************************************************************************************
// ***********************************************************************************************

// CMD command
type CMD struct {
	PatternType          string
	Pattern              []string
	Description          string
	Required             bool
	compiledRegex        []*regexp.Regexp
	Handler              CMDHandler
	HandlerError         CMDHandler
	HandlerParams        CMDHandler
	HandlerParamsMissing CMDHandler
	HandlerParamsExtra   CMDHandler
	SubCommands          []CMD
	Params               []CMDParam
}

// AddSubCMD adds a sub command
func (st *CMD) AddSubCMD(sub CMD) {
	st.SubCommands = append(st.SubCommands, sub)
}

// addCompiledRegex adds a new compiled regex (that matches a pattern)
func (st *CMD) addCompiledRegex(regx *regexp.Regexp) {
	if st.compiledRegex == nil {
		st.compiledRegex = make([]*regexp.Regexp, 0)
	}

	st.compiledRegex = append(st.compiledRegex, regx)
}

// ***********************************************************************************************
// **  CMDParam  *********************************************************************************
// ***********************************************************************************************

type CMDParam struct {
	ID          string
	Description string
	Required    bool
	Value       string
}

// GetIntValue return the int value
func (st *CMDParam) GetIntValue() (int, error) {
	return strconv.Atoi(st.Value)
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
