package goflowdock

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	CMDTypeRegex = "regex"
	CMDTypeWord  = "word"

	CMDHandlerTypeDefault         = "handler.default"
	CMDHandlerTypeError           = "handler.error"
	CMDHandlerTypeParams          = "handler.params"
	CMDHandlerTypeRestrictions    = "handler.restrictions"
	CMDHandlerTypeParamsWrongType = "handler.params.wrong.type"
	CMDHandlerTypeParamsMissing   = "handler.params.missing"
	CMDHandlerTypeParamsExtra     = "handler.params.extra"

	CMDExtraDataParametersIncorrectType = "extra.data.parameters.incorrect.type"
	CMDExtraDataParametersMissing       = "extra.data.parameters.missing"
	CMDExtraDataParametersExtra         = "extra.data.parameters.extra"

	// CMD flow restriction, includes only the listed flows
	CMDRestrictionFlowConceptInclude = "restriction.flow.include"
	// CMD flow restriction, excludes the listed flows, allows all other flows
	CMDRestrictionFlowConceptExclude = "restriction.flow.exclude"
)

// CMDHandler is the function handler
//
// Parameters:
//
// cmd CMD				==> CMD that matched
// pattern string		==> exact CMD pattern that matched
// entry Entry			==> Related Flowdock entry
// cmdContent string	==> Content to parse (original content - cmd)
// extraData			==> Extra data related to the CMD
// handlerType string	==> Indicates which handler type will be used and gives some extra data (parameters: ok, missing, extra)
type CMDHandler func(cmd CMD, pattern string, entry Entry, cmdContent string, extraData map[string]interface{}, handlerType string)

// ***********************************************************************************************
// **  CMD  **************************************************************************************
// ***********************************************************************************************

// CMD command
type CMD struct {
	PatternType            string
	Pattern                []string
	Description            string
	Required               bool
	compiledRegex          []*regexp.Regexp
	Handler                CMDHandler
	HandlerError           CMDHandler
	HandlerRestrictions    CMDHandler
	HandlerParams          CMDHandler
	HandlerParamsWrongType CMDHandler
	HandlerParamsMissing   CMDHandler
	HandlerParamsExtra     CMDHandler
	SubCommands            []CMD
	Params                 []CMDParam
	params2map             bool
	mpParams               map[string]CMDParam
	Restrictions           CMDRestriction
}

type CMDRestriction struct {
	Flows []CMDRestrictionFlow
}

type CMDRestrictionFlow struct {
	Concept string
	Data    []CMDRestrictionFlowData
}

type CMDRestrictionFlowData struct {
	ID   string
	Name string
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

// GetParamByID returns a CMDParam given its ID
func (st *CMD) GetParamByID(id string) (CMDParam, error) {
	if !st.params2map {
		st.convertParamsToMap()
	}

	if param, ok := st.mpParams[id]; ok {
		return param, nil
	}

	return CMDParam{}, fmt.Errorf("No param '%v'", id)
}

// GetParamByIndex returns a CMDParam given its index
func (st *CMD) GetParamByIndex(index int) (CMDParam, error) {
	if index >= len(st.Params) {
		return CMDParam{}, fmt.Errorf("Index out of bounds: %v", index)
	}

	return st.Params[index], nil
}

// convertParamsToMap converts Params ([]CMDParam) to a map[string]CMDParam
func (st *CMD) convertParamsToMap() {
	st.mpParams = make(map[string]CMDParam)
	for i := 0; i < len(st.Params); i++ {
		st.mpParams[st.Params[i].ID] = st.Params[i]
	}

	st.params2map = true
}

// GetReady get the CMD ready:
// 	- Iterate over the Restrictions and fill all missing data (if possible)
func (st *CMD) GetReady(flowManager *FlowManager) {
	st.getReadyRestrictionsFlows(flowManager)
}

// getReadyRestrictionsFlows iterates over the flow's restrictions and fills the flow's IDS (if possible)
func (st *CMD) getReadyRestrictionsFlows(flowManager *FlowManager) {
	// get ready all the CMD's flows
	for i := 0; i < len(st.Restrictions.Flows); i++ {
		for c := 0; c < len(st.Restrictions.Flows[i].Data); c++ {
			if st.Restrictions.Flows[i].Data[c].ID == "" && st.Restrictions.Flows[i].Data[c].Name != "" {
				// get the flow ID
				// TODO ::: cache && use cached results for the following search
				if flow, err := flowManager.GetByName(st.Restrictions.Flows[i].Data[c].Name); err == nil {
					st.Restrictions.Flows[i].Data[c].ID = flow.ID
				}
			}
		}
	}

	// get ready all the CMD's subCMDs
	for i := 0; i < len(st.SubCommands); i++ {
		st.SubCommands[i].GetReady(flowManager)
	}

}

// CanExecute returns true whether the CMD can be executed. It verifies all restrictions.
func (st *CMD) CanExecute(entry Entry) (bool, error) {
	canExecute := true

	// check flows restrictions
	canExecute = st.canExecuteByFlows(entry)

	return canExecute, nil
}

// canExecuteByFlows verifies all Flow restrictions
func (st *CMD) canExecuteByFlows(entry Entry) bool {
	canExecute := true

	if len(st.Restrictions.Flows) > 0 {
		for i := 0; i < len(st.Restrictions.Flows); i++ {
			switch st.Restrictions.Flows[i].Concept {
			case CMDRestrictionFlowConceptInclude:
				canExecute = false
				for c := 0; c < len(st.Restrictions.Flows[i].Data); c++ {
					if entry.Flow == st.Restrictions.Flows[i].Data[c].ID {
						canExecute = true
						break
					}
				}

			case CMDRestrictionFlowConceptExclude:
				canExecute = true
				for c := 0; c < len(st.Restrictions.Flows[i].Data); c++ {
					if entry.Flow == st.Restrictions.Flows[i].Data[c].ID {
						canExecute = false

						// TODO ::: Return a text message saying why the CMD cannot be executed

						break
					}
				}
			}
		}
	}

	return canExecute
}

// ***********************************************************************************************
// **  CMDParam  *********************************************************************************
// ***********************************************************************************************

const (
	CMDParamTypeInt = "int"
)

type CMDParam struct {
	ID          string
	Description string
	Type        string
	Required    bool
	Value       string
}

// isExpectedType verifies whether the param's type is the expected.
func (st *CMDParam) isExpectedType() bool {
	switch st.Type {
	case "":
		return true

	case CMDParamTypeInt:
		_, err := st.GetIntValue()
		return err == nil
	}

	return false
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
