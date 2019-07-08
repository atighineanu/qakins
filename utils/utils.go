package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

type Upd struct {
	Inc  string
	Name []string
	Chan map[string]string
}

type SSHInfo struct {
	User string
	Pass string
	IP   string
}

type PipelineCfg struct {
	Username    string
	Password    string
	DockerRepo  string
	PackageName string
	PipeName    string
}

func (s *SSHInfo) Command(cmd ...string) *exec.Cmd {

	arg := append(
		[]string{"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", s.User, s.IP),
		},
		cmd...,
	)
	return exec.Command("ssh", arg...)
}

func SSHPrinter(command []string, user SSHInfo) string {

	out, err := user.Command(command...).CombinedOutput()
	if err != nil {
		fmt.Printf("Bad! Error at runng osc thourgh SSH... %s\n", err)
	}
	return fmt.Sprintf("%s", string(out))
}

//fly -t tutorial sp -p NEW_PIPELINE -c pipeline.yml -v repo=http://download.suse.de/ibs/SUSE:/Maintenance:/11566/SUSE_Updates_SUSE-CAASP_3.0_x86_64/SUSE:Maintenance:11566.repo -n

func ConcourseRunner(Repo string, Incident Incident, config PipelineCfg) *exec.Cmd {
	var pipename string
	if config.PipeName != "" {
		pipename = config.PipeName
	} else {
		pipename = fmt.Sprintf("%s %s %s", config.PackageName, Incident.Base.Project, Incident.Update.Severity)
		pipename = strings.Replace(pipename, "/", "-", 10)
	}
	arg := []string{"-t", "tutorial", "sp", "-p", pipename, "-c", "../main_pipeline.yml", "\\",
		"-v", fmt.Sprintf("repo=%s", Repo), "-n", "\\",
		"-v", fmt.Sprintf("user=%s", config.Username), "-n", "\\",
		"-v", fmt.Sprintf("password=%s", config.Password), "-n", "\\",
		"-v", fmt.Sprintf("package=%s", config.PackageName), "-n", "\\",
		"-v", fmt.Sprintf("docker_repository=%s", config.DockerRepo), "-n", "\\"}
	out := exec.Command("fly", arg...)
	return out
}
