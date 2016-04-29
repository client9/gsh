package gsh

import (
	"flag"
	"fmt"
	"time"

	"gopkg.in/pipe.v2"
)

type ParseTimeCmd struct {
	InFormat  string
	OutFormat string
}

func (cmd *ParseTimeCmd) Name() string {
	return "parseTime"
}
func (cmd *ParseTimeCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	f.StringVar(&cmd.InFormat, "in", "", "golang time format string")
	f.StringVar(&cmd.OutFormat, "out", "RFC3339", "golang time format string")
	return f
}

func (cmd *ParseTimeCmd) convert(line []byte) ([]byte, error) {
	t, err := time.Parse(cmd.InFormat, string(line))
	if err != nil {
		return nil, fmt.Errorf("%s: unable to parse %q with %q", cmd.Name(), line, cmd.InFormat)
	}
	return []byte(t.UTC().Format(cmd.OutFormat)), nil
}

func (cmd *ParseTimeCmd) Run(s *pipe.State, argv []string) error {
	fs := cmd.Flags()
	err := fs.Parse(argv)
	if err != nil {
		return err
	}
	if len(cmd.InFormat) == 0 {
		return fmt.Errorf("%s: Must specify format with -f", cmd.Name)
	}
	if cmd.InFormat == "RFC1123Z" {
		cmd.InFormat = time.RFC1123Z
	}
	if cmd.OutFormat == "RFC3339" {
		cmd.OutFormat = time.RFC3339
	}
	args := fs.Args()
	return ForEachLine(s, args, cmd.convert)
}

func ParseTime(argv ...string) pipe.Pipe {
	cmd := ParseTimeCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}
