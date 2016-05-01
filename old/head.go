package gsh

import (
	"flag"
	"fmt"
	"io"
	"os"

	"gopkg.in/pipe.v2"
)

type HeadCmd struct {
	Chars int
	Lines int
}

func (cmd *HeadCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	f.IntVar(&cmd.Chars, "c", -1, "Take first N characters")
	f.IntVar(&cmd.Chars, "n", 10, "Take first N lines")
	return f
}

func (cmd *HeadCmd) Name() string {
	return "head"
}

func (cmd *HeadCmd) Run(s *pipe.State, argv []string) error {
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
			if err != nil {
				return fmt.Errorf("%s 2: %s", cmd.Name(), err)
			}
		}
		return nil
	}
	if cmd.Chars >= 0 {
		_, err = io.CopyN(s.Stdout, s.Stdin, int64(cmd.Chars))
	}
	if err != nil {
		return fmt.Errorf("%s: %s", cmd.Name(), err)
	}
	return nil
}

func Head(argv ...string) pipe.Pipe {
	cmd := HeadCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}
