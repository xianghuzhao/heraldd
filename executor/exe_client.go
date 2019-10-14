package executor

import (
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/xianghuzhao/heraldd/util"
)

// ExeClient is an executor which runs jobs locally
type ExeClient struct {
	util.BaseLogger
	WorkDir string
}

// Execute will run job locally
func (exe *ExeClient) Execute(param map[string]interface{}) map[string]interface{} {
	err := os.MkdirAll(exe.WorkDir, 0755)
	if err != nil {
		exe.Logger.Errorf("[Executor(ExeClient)] Create work directory \"%s\" failed: %s", exe.WorkDir, err)
	}

	scriptRepo, _ := util.GetStringParam(param, "script_repo")
	scriptBranch, _ := util.GetStringParam(param, "script_branch")
	scriptCommand, _ := util.GetStringParam(param, "command")

	repoParsed, err := url.Parse(scriptRepo)
	if err != nil {
		exe.Logger.Errorf("[Executor(ExeClient)] Invalid repo name: %s", scriptRepo)
		return nil
	}

	host := strings.SplitN(repoParsed.Host, ":", 2)[0]
	repoPathFrags := []string{exe.WorkDir, "repo", host}
	urlPath := strings.TrimLeft(repoParsed.Path, "/")
	if strings.HasSuffix(urlPath, ".git") {
		urlPath = urlPath[:len(urlPath)-4]
	}
	repoPathFrags = append(repoPathFrags, strings.Split(urlPath, "/")...)
	repoPath := path.Join(repoPathFrags...)

	if stat, err := os.Stat(repoPath); os.IsNotExist(err) {
		err := RunCmd([]string{"git", "clone", scriptRepo, repoPath}, "")
		if err != nil {
			exe.Logger.Errorf("[Executor(ExeClient)] %s", err)
			return nil
		}
	} else {
		if !stat.IsDir() {
			exe.Logger.Errorf("[Executor(ExeClient)] Path for repo is not a directory: %s", repoPath)
			return nil
		}
		err := RunCmd([]string{"git", "fetch", "--all"}, repoPath)
		if err != nil {
			exe.Logger.Errorf("[Executor(ExeClient)] %s", err)
			return nil
		}
	}

	if scriptBranch == "" {
		scriptBranch = "master"
	}
	err = RunCmd([]string{"git", "checkout", "refs/remotes/origin/" + scriptBranch}, repoPath)
	if err != nil {
		exe.Logger.Errorf("[Executor(ExeClient)] %s", err)
		return nil
	}

	runDir := path.Join(exe.WorkDir, "run")
	err = os.MkdirAll(runDir, 0755)
	if err != nil {
		exe.Logger.Errorf("[Executor(ExeClient)] Create run directory \"%s\" failed: %s", runDir, err)
	}

	err = RunCmd([]string{path.Join(repoPath, scriptCommand)}, runDir)
	if err != nil {
		exe.Logger.Errorf("[Executor(ExeClient)] %s", err)
		return nil
	}

	return map[string]interface{}{
		"print": true,
	}
}

// SetParam will set param from a map
func (exe *ExeClient) SetParam(param map[string]interface{}) {
	util.UpdateStringParam(&exe.WorkDir, param, "work_dir")
}
