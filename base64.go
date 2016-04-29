package gsh

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"

	"gopkg.in/pipe.v2"
)

type Base64Cmd struct {
}

func (cmd *Base64Cmd) Name() string {
	return "base64"
}
func (cmd *Base64Cmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	return f
}

func (cmd *Base64Cmd) Run(s *pipe.State, argv []string) error {
	fs := cmd.Flags()
	err := fs.Parse(argv)
	if err != nil {
		return err
	}
	stdin := base64.NewEncoder(base64.StdEncoding, s.Stdout)
	_, err = io.Copy(stdin, s.Stdin)
	if err != nil {
		return fmt.Errorf("%s: %s", cmd.Name(), err)
	}
	return nil
}

func Base64(argv ...string) pipe.Pipe {
	cmd := Base64Cmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}
