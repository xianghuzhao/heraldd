package selector

import (
	"encoding/json"

	"github.com/heraldgo/heraldd/util"
)

// External is a selector call a sub process to check the result
type External struct {
	util.BaseLogger
	Program string
}

// Select will call a sub process to check the exit code
func (slt *External) Select(triggerParam, selectParam map[string]interface{}) bool {
	triggerParamJSON, err := json.Marshal(triggerParam)
	if err != nil {
		slt.Errorf("Generate trigger param argument failed: %s", err)
		return false
	}

	selectParamJSON, err := json.Marshal(selectParam)
	if err != nil {
		slt.Errorf("Generate selector param argument failed: %s", err)
		return false
	}

	env := []string{
		"HERALD_TRIGGER_PARAM=" + string(triggerParamJSON),
		"HERALD_SELECT_PARAM=" + string(selectParamJSON),
	}

	exitCode, err := util.RunCmd([]string{slt.Program}, "", env, false, nil, nil)
	if err != nil {
		slt.Errorf("Run external selector error: %s", err)
		return false
	}
	if exitCode != 0 {
		slt.Debugf("External selector does not pass: exit(%d), err(%s)", exitCode, err)
		return false
	}

	return true
}

func newSelectorExternal(param map[string]interface{}) interface{} {
	program, _ := util.GetStringParam(param, "program")

	return &External{
		Program: program,
	}
}
