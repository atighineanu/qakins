package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

var (
	Jobs = []string{"ScenarioI(Updating_alltogether_SCC_TestPackage)", "ScenarioII(Updating_separately_1stSCC_2ndTestPackage)"}
)

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

func ConcourseRunner(Repo string, Incident Incident, config PipelineCfg) (*exec.Cmd, string) {
	var pipename string
	if config.PipeName != "" {
		pipename = config.PipeName
	} else {
		pipename = fmt.Sprintf("%s_%s_%s", config.PackageName, Incident.Base.Project, Incident.Update.Severity)
		pipename = strings.Replace(pipename, "/", "-", 10)
	}
	arg := []string{"-t", "tutorial", "sp", "-p", pipename, "-c", "../main_pipeline.yml", "\\",
		"-v", fmt.Sprintf("repo=%s", Repo), "-n", "\\",
		"-v", fmt.Sprintf("user=%s", config.Username), "-n", "\\",
		"-v", fmt.Sprintf("password=%s", config.Password), "-n", "\\",
		"-v", fmt.Sprintf("package=%s", config.PackageName), "-n", "\\",
		"-v", fmt.Sprintf("docker_repository=%s", config.DockerRepo), "-n", "\\"}
	out := exec.Command("fly", arg...)
	return out, pipename
}

func FlyLogin() {
	cmdargs := []string{"fly", "login", "-u", "test", "-p", "test", "-t", "tutorial"}
	_, err := exec.Command(cmdargs[0], cmdargs[1:]...).CombinedOutput()
	if err != nil {
		log.Printf("Error occured: %s\n", err)
	}

	cmdargs = []string{"fly", "status", "-t", "tutorial"}
	out, err := exec.Command(cmdargs[0], cmdargs[1:]...).CombinedOutput()
	if err != nil {
		log.Printf("Error occured: %s\n", err)
	}
	if strings.Contains(fmt.Sprintf("%s", string(out)), "successfully") {
		log.Println("logged into concourse successfully...")
	} else {
		log.Println("login into concourse failed!")
	}
}

func FlyJobTrigg(pipeline string) error {
	//--------------------Check if Pipeline Exists-----------------
	cmdargs := []string{"fly", "-t", "tutorial", "pipelines"}
	out, err := exec.Command(cmdargs[0], cmdargs[1:]...).CombinedOutput()
	if err != nil {
		log.Printf("Error occured: %s\n", err)
		return err
	}
	if strings.Contains(fmt.Sprintf("%s", string(out)), pipeline) {
		log.Println("Pipeline % was properly set...", pipeline)
	}
	//---------------------Unpause Pipeline-------------------------
	cmdargs = []string{"fly", "-t", "tutorial", "unpause-pipeline", "-p", pipeline}
	out, err = exec.Command(cmdargs[0], cmdargs[1:]...).CombinedOutput()
	if err != nil {
		log.Printf("Error occured: %s\n", err)
		return err
	}
	if strings.Contains(fmt.Sprintf("%s", string(out)), "unpaused") {
		log.Println("Pipeline %s was successfully unpaused...", pipeline)
	}
	//-----------------------Trigger Jobs--------------------------
	//fly -t tutorial trigger-job -j "kubernetes-salt SUSE:Maintenance:11964 moderate/ScenarioI(Updating all at once: SCC and TestPackage)"
	for _, job := range Jobs {
		cmdargs = []string{"fly", "-t", "tutorial", "trigger-job", "-j", filepath.Join(pipeline, job)}
		out, err = exec.Command(cmdargs[0], cmdargs[1:]...).CombinedOutput()
		if err != nil {
			log.Printf("Error occured: %s\n", err)
		}
		err := CheckIfDone(pipeline, job)
		if err != nil {
			log.Printf("Error occured while running CheckIfDoneFunction: %s\n", err)
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}

func CheckIfDone(pipeline string, job string) error {
	breakcycle := false
	t := time.Now()
	cmdargs := []string{"fly", "-t", "tutorial", "jobs", "-p", pipeline}
	time.Sleep(10 * time.Second)
	for {
		out, err := exec.Command(cmdargs[0], cmdargs[1:]...).CombinedOutput()
		if err != nil {
			log.Printf("Error found: %s", err)
			return err
		}
		temp := strings.Split(fmt.Sprintf("%s", string(out)), "\n")
		for _, k := range temp {
			if strings.Contains(k, job) && !strings.Contains(k, "started") {
				log.Printf("The job: %s->%s Is successfully done! (log.time= %2.2f seconds)", pipeline, job, time.Since(t).Seconds())
				breakcycle = true
				break
				break
			}
		}
		if breakcycle == true {
			break
		}
		time.Sleep(60 * time.Second)
		log.Printf("Running %s->%s Job for %2.2f seconds now...", pipeline, job, time.Since(t).Seconds())
	}
	return nil
}

func NiceBuffRunner(cmd *exec.Cmd, workdir string) (string, string) {
	var stdoutBuf, stderrBuf bytes.Buffer
	//newEnv := append(os.Environ(), ENV...)
	//cmd.Env = newEnv
	cmd.Dir = workdir
	pipe, _ := cmd.StdoutPipe()
	errpipe, _ := cmd.StderrPipe()
	var errStdout, errStderr error
	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)
	err := cmd.Start()
	if err != nil {
		return fmt.Sprintf("%s", os.Stdout), fmt.Sprintf("%s", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		_, errStdout = io.Copy(stdout, pipe)
		wg.Done()
	}()
	go func() {
		_, errStderr = io.Copy(stderr, errpipe)
		wg.Wait()
	}()
	err = cmd.Wait()
	if err != nil {
		return fmt.Sprintf("%s", os.Stdout), fmt.Sprintf("%s", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("Command runninng error: failed to capture stdout or stderr\n")
	}
	return stdoutBuf.String(), stderrBuf.String()
}
