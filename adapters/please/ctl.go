package please

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/tcncloud/wollemi/ports/please"
)

func NewCtl() *Ctl {
	return &Ctl{}
}

type Ctl struct{}

func (*Ctl) QueryDeps(targets ...string) ([]string, error) {
	args := make([]string, 0, 4+len(targets))
	args = append(args, "query", "deps", "-h")
	args = append(args, targets...)

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	cmd := exec.Command("plz", args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf(stderr.String())
	}

	r := bufio.NewReader(stdout)

	var deps []string

	have := make(map[string]bool)

	for {
		dep, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		dep = strings.TrimSpace(dep)

		if !have[dep] {
			deps = append(deps, dep)
			have[dep] = true
		}
	}

	return deps, nil
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
