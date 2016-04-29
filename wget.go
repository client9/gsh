package gsh

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"gopkg.in/pipe.v2"
)

type WgetCmd struct {
	OutputFile string
	Method     string
}

func (cmd *WgetCmd) Name() string {
	return "wget"
}

func (cmd *WgetCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	f.StringVar(&cmd.Method, "method", "GET", "HTTP Method")
	f.StringVar(&cmd.OutputFile, "O", "", "Output file, base of URL")
	return f
}

func (cmd *WgetCmd) Run(s *pipe.State, argv []string) error {
	fs := cmd.Flags()
	err := fs.Parse(argv)
	if err != nil {
		return err
	}
	args := fs.Args()
	if len(args) != 1 {
		return fmt.Errorf("%s: expected only 1 URL", cmd.Name())
	}
	source := args[0]

	client := &http.Client{}
	req, err := http.NewRequest(cmd.Method, source, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to create request: %s", cmd.Name(), err)
	}
	// req.Header.Add("If-None-Match", `W/"wyzzy"`)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: request failed: %s", cmd.Name(), err)
	}
	var out io.Writer
	if cmd.OutputFile == "-" {
		out = s.Stdout
	} else {
		if len(cmd.OutputFile) == 0 {
			cmd.OutputFile = path.Base(source)
		}
		out, err = os.Create(cmd.OutputFile)
		if err != nil {
			return fmt.Errorf("%s: request failed %s", cmd.Name(), err)
		}
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("%s: request failed %s", cmd.Name(), err)
	}
	return nil
}


func Wget(argv ...string) pipe.Pipe {
	cmd := WgetCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}

type HTTPLastModifiedCmd struct {
}

func (cmd *HTTPLastModifiedCmd) Name() string {
	return "http_last_mod"
}

func (cmd *HTTPLastModifiedCmd) Flags() *flag.FlagSet {
	f := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	return f
}

func (cmd *HTTPLastModifiedCmd) Run(s *pipe.State, argv []string) error {
	fs := cmd.Flags()
	err := fs.Parse(argv)
	if err != nil {
		return err
	}
	args := fs.Args()
	if len(args) != 1 {
		return fmt.Errorf("%s: expected only 1 URL", cmd.Name())
	}
	source := args[0]

	resp, err := http.Head(source)
	if err != nil {
		return fmt.Errorf("%s: failed to create request: %s", cmd.Name(), err)
	}
	resp.Body.Close()
	lastModified :=  resp.Header.Get("Last-Modified")
	if len(lastModified) == 0 {
		return fmt.Errorf("%s: no Last-Modified field")
	}
	tobj, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		return fmt.Errorf("%s: request failed %s", cmd.Name(), err)
	}
	s.Stdout.Write([]byte(tobj.UTC().Format(time.RFC3339)))
	s.Stdout.Write([]byte{'\n'})
	return nil

}

func HTTPLastModified(argv ...string) pipe.Pipe {
	cmd := HTTPLastModifiedCmd{}
	return pipe.TaskFunc(func(s *pipe.State) error {
		return cmd.Run(s, argv)
	})
}
