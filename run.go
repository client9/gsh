package gsh

import (
	"github.com/google/shlex"
	"gopkg.in/pipe.v2"
)

var CmdMap = map[string]func(argv ...string) pipe.Pipe{
	"wget":             Wget,
	"which":            Which,
	"base64":           Base64,
	"cat":              Cat,
	"head":             Head,
	"gitLastModified":  GitLastModified,
	"fileLastModified": FileLastModified,
	"strptime":         ParseTime,
}

// Run runs a full shell style command "ls -l /foo"
func Run(argv ...string) pipe.Pipe {
	if len(argv) == 0 {
		return nil
	}
	args, err := shlex.Split(argv[0])
	if err != nil {
		return nil
	}

	// now group based on "|" command
	pipes := make([]pipe.Pipe, 0)
	cmdarg := make([]string, 0)

	for _, a := range args {
		if a == "|" {
			p := RunArgs(cmdarg[0], cmdarg[1:]...)
			if p == nil {
				return nil
			}
			pipes = append(pipes, p)
			cmdarg = make([]string, 0)
		} else {
			cmdarg = append(cmdarg, a)
		}
	}
	if len(cmdarg) > 0 {
		p := RunArgs(cmdarg[0], cmdarg[1:]...)
		if p == nil {
			return nil
		}
		pipes = append(pipes, p)
	}

	// minor optimization if only 1
	if len(pipes) == 1 {
		return pipes[0]
	}

	return pipe.Line(pipes...)
}

// RunArgs runs a single command with explicit arguments
func RunArgs(cmd string, args ...string) pipe.Pipe {
	fn, ok := CmdMap[cmd]
	if !ok {
		return pipe.Exec(cmd, args...)
	}
	return fn(args...)
}
