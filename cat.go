package gsh

import (
	"flag"
	"fmt"
	"io"
	"os"

	"gopkg.in/pipe.v2"
)

func CatShowVisible() string {
	return "-v"
}

type CatCmd struct {
	ShowVisible bool
}

func (cmd *CatCmd) Name() string {
	return "cat"
}
func (cmd *CatCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	f.BoolVar(&cmd.ShowVisible, "v", true, "Show invisible charcters")
	return f
}

func (cmd *CatCmd) Run(s *pipe.State, argv []string) error {
	fs := cmd.Flags()
	err := fs.Parse(argv)
	if err != nil {
		return err
	}
	args := fs.Args()
	if len(args) > 0 {
		for _, fname := range args {
			fh, err := os.Open(fname)
			if err != nil {
				return fmt.Errorf("%s 1: %s", cmd.Name(), err)
			}
			_, err = io.Copy(s.Stdout, fh)
			if err != nil && err != io.ErrClosedPipe {
				return fmt.Errorf("%s 2: %+v", cmd.Name(), err)
			}
		}
		return nil
	}
	_, err = io.Copy(s.Stdout, s.Stdin)
	if err != nil {
		return fmt.Errorf("%s 3: %s", cmd.Name(), err)
	}
	return nil
}

func Cat(argv ...string) pipe.Pipe {
	cmd := CatCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}
