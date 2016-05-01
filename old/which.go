package gsh

import (
	"flag"
	"fmt"
	"os/exec"

	"gopkg.in/pipe.v2"
)

type WhichCmd struct {
}

func (cmd *WhichCmd) Name() string {
	return "which"
}

func (cmd *WhichCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	return f
}

func (cmd *WhichCmd) Run(s *pipe.State, args []string) error {
	return nil
}

func (cmd *WhichCmd) which(line []byte) ([]byte, error) {
	exe, err := exec.LookPath(string(line))
	if err != nil {
		// nothing found, not an error
		return nil, nil
	}
	return []byte(exe), nil

}

// Which
func Which(argv ...string) pipe.Pipe {
	cmd := WhichCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		err := ForEachLine(s, argv, cmd.which)
		if err != nil {
			return fmt.Errorf("%s: failed: %s", cmd.Name(), err)
		}
		return nil
	})
}
