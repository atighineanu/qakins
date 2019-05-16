package utils

import (
	"fmt"
	"os/exec"
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
