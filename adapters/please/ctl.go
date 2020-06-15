package please

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/tcncloud/wollemi/ports/please"
)

func NewCtl() *Ctl {
	return &Ctl{}
}

type Ctl struct{}

func (*Ctl) QueryDeps(targets ...string) (*bufio.Reader, error) {
	args := make([]string, 0, 4+len(targets))
	args = append(args, "query", "deps", "-u", "-h")
	args = append(args, targets...)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd := exec.Command("plz", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(stderr.String())
	}

	return bufio.NewReader(stdout), nil
}

func (*Ctl) Build(targets ...string) error {
	args := make([]string, 0, 2+len(targets))
	args = append(args, "build", "-p")
	args = append(args, targets...)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd := exec.Command("plz", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(stderr.String())
	}

	return nil
}

func (*Ctl) Graph() (*please.Graph, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd := exec.Command("plz", "query", "graph")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(stderr.String())
	}

	graph := &please.Graph{}

	if err := json.Unmarshal(stdout.Bytes(), graph); err != nil {
		return nil, err
	}

	return graph, nil
}
