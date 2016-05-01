package gsh

import (
	"flag"
	"os"
	"time"

	"gopkg.in/pipe.v2"
)

type GitLastModifiedCmd struct {
}

func (cmd *GitLastModifiedCmd) Name() string {
	return "gitLastModified"
}

func (cmd *GitLastModifiedCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	return f
}

func (cmd *GitLastModifiedCmd) Run(s *pipe.State, args []string) error {
	return nil
}

// GitLastModified returns the the last-modified timestamp of a given file or files
func GitLastModified(argv ...string) pipe.Pipe {
	//cmd := GitLastModifiedCmd{}
	gitargs := []string{"log", "-n", "1", "--date=rfc", "--pretty=format:%cd"}
	gitargs = append(gitargs, argv...)

	return pipe.Line(
		pipe.Exec("git", gitargs...),
		ParseTime("-in", "RFC1123Z"),
	)
}

type FileLastModifiedCmd struct {
}

func (cmd *FileLastModifiedCmd) Name() string {
	return "fileLastModified"
}

func (cmd *FileLastModifiedCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	return f
}

func (cmd *FileLastModifiedCmd) exec(line []byte) ([]byte, error) {
	fi, err := os.Stat(string(line))
	if err != nil {
		return nil, err
	}
	return []byte(fi.ModTime().Format(time.RFC3339)), nil
}

func (cmd *FileLastModifiedCmd) Run(s *pipe.State, argv []string) error {
	fs := cmd.Flags()
	err := fs.Parse(argv)
	if err != nil {
		return err
	}
	args := fs.Args()
	return ForEachLine(s, args, cmd.exec)
}
func FileLastModified(argv ...string) pipe.Pipe {
	cmd := FileLastModifiedCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}

/*
func DirLastModified(argv ...string) pipe.Pipe {

}

*/
