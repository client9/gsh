package gsh

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

type FuncMap map[string](func(*Session, []string) error)

func New() *Session {
	s := Session{}
	s.Env = envMap(os.Environ())
	s.alias = make(map[string][]string)
	s.fmap = FuncMap{
		"alias":   Alias,
		"cp":      Copy,
		"echo":    Echo,
		"export":  Export,
		"mkdir":   Mkdir,
		"mv":      Move,
		"unalias": Unalias,
		"which":   Which,
	}
	return &s
}

func (s *Session) GetEnv(key string) string {
	return s.Env[key]
}

func (s *Session) PutEnv(key string, val string) {
	s.Env[key] = val
}

func (s *Session) SetError(e error) {
	s.err = e
}
func (s *Session) Error() error {
	return s.err
}

func (s *Session) Funcs(funcs FuncMap) *Session {
	for k, v := range funcs {
		if v == nil {
			delete(s.fmap, k)
		} else {
			s.fmap[k] = v
		}
	}
	return s
}

func (s *Session) Script(str string) *Session {
	// break apart script into single commands
	s.cmds = unbreak(str)
	return s
}

func (s *Session) Output() ([]byte, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	s.Stdout = &stdout
	s.Stderr = &stderr

	err := s.Run()
	if err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}

func (s *Session) Exec(cmds ...string) error {
	if s.Error() != nil {
		return nil
	}
	switch len(cmds) {
	case 0:
		return fmt.Errorf("Exec called without args?")
	case 1:
		s.cmds = unbreak(cmds[0])
	default:
		s.cmds = cmds
	}
	return s.Run()
}

func (s *Session) Run() error {
	if s.Error() != nil {
		return nil
	}
	for _, cmd := range s.cmds {
		// replace all ${  } shell variables
		cmd = os.Expand(cmd,
			(func(key string) string { return s.Env[key] }))

		parts, err := shlex.Split(cmd)
		if err != nil {
			s.SetError(fmt.Errorf("Bad line %q: %s", cmd, err))
			return err
		}
		// effectively blank line
		if len(parts) == 0 {
			continue
		}

		log.Printf("RUNNING: %s", cmd)

		// TODO: is arg0 an environment override
		// only for external commands I think

		// Insert Alias
		if newargs, ok := s.alias[parts[0]]; ok {
			parts = append(newargs, parts[1:]...)
			log.Printf("ALIAS: %s", strings.Join(parts, " "))
		}

		fn, ok := s.fmap[parts[0]]
		if ok {
			err := fn(s, parts)
			if err != nil {
				s.SetError(err)
				return err
			}
			continue
		}
		log.Printf("Shelling out... not in map: %s", parts[0])
		// ok shell out
		execCmd := exec.Command(parts[0], parts[1:]...)

		// set up environment
		execCmd.Stdin = nil
		execCmd.Stdout = s.Stdout
		execCmd.Stderr = s.Stderr

		// TODO DIR/PATH
		// TODO ENV

		err = execCmd.Run()
		if err != nil {
			s.SetError(err)
			return err
		}
	}
	return nil
}

func Export(s *Session, cli []string) error {
	if len(cli) != 2 {
		return fmt.Errorf("%s: Expected only 1 arg: got %v", cli[0], cli[1:])
	}
	//name := cli[0]
	kv := cli[1]
	idx := strings.IndexByte(kv, '=')
	if idx == -1 {
		return fmt.Errorf("didnt find key/value")
	}
	s.PutEnv(kv[:idx], kv[idx+1:])
	return nil
}

func Echo(s *Session, cli []string) error {
	//name := cli[0]
	fargs := cli[1:]

	// no flags

	s.Stdout.Write([]byte(strings.Join(fargs, " ")))

	return nil
}

func Which(s *Session, cli []string) error {
	//name := cli[0]
	fargs := cli[1:]

	// no flags
	cmd, err := exec.LookPath(fargs[0])
	if err != nil {
		return err
	}
	s.Stdout.Write([]byte(cmd))
	return nil
}

func Mkdir(s *Session, cli []string) error {
	name := cli[0]
	fargs := cli[1:]
	var parents bool
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.BoolVar(&parents, "p", false, "create parent directories")
	err := f.Parse(fargs)
	if err != nil {
		return err
	}
	for _, dirs := range f.Args() {
		if parents {
			err = os.MkdirAll(dirs, 0777)
		} else {
			err = os.Mkdir(dirs, 0777)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func Alias(s *Session, cli []string) error {
	name, fargs := cli[0], cli[1:]
	f := flag.NewFlagSet(name, flag.ExitOnError)
	for err := f.Parse(fargs); err != nil; {
		return err
	}
	args := f.Args()
	switch len(args) {
	case 0:
		return fmt.Errorf("%s requires at least one arg", name)
	case 1:
		aliasargs, ok := s.alias[args[0]]
		if !ok {
			return fmt.Errorf("%s: not found %s", name, fargs[0])
		}
		s.Stdout.Write([]byte(strings.Join(aliasargs, " ")))
		s.Stdout.Write([]byte("\n"))
		return nil
	default:
		s.alias[args[0]] = args[1:]
		return nil
	}
}

func Unalias(s *Session, cli []string) error {
	name, fargs := cli[0], cli[1:]
	f := flag.NewFlagSet(name, flag.ExitOnError)
	err := f.Parse(fargs)
	if err != nil {
		return err
	}
	args := f.Args()
	switch len(args) {
	case 0:
		return fmt.Errorf("%s requires at least one arg", name)
	case 1:
		delete(s.alias, args[0])
		return nil
	default:
		return fmt.Errorf("%s too many args", name)
	}
}

type Session struct {
	err   error
	alias map[string][]string
	fmap  map[string](func(*Session, []string) error)
	cmds  []string

	Env    map[string]string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// envMap converts an environment in []string{"k=v"}
// map[k] = v
func envMap(orig []string) map[string]string {
	out := make(map[string]string)
	for _, kv := range orig {
		idx := strings.IndexByte(kv, '=')
		if idx != -1 {
			key := kv[:idx]
			val := kv[idx+1:]
			out[key] = val
		}
	}
	log.Printf("DONE")
	return out
}

func newEnviron(env map[string]string, inherit bool) []string { //map[string]string {
	environ := make([]string, 0, len(env))
	if inherit {
		for _, line := range os.Environ() {
			for k := range env {
				if strings.HasPrefix(line, k+"=") {
					goto CONTINUE
				}
			}
			environ = append(environ, line)
		CONTINUE:
		}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	return environ
}

// unbreak takes a lines and
func unbreak(s string) []string {
	lines := strings.Split(s, "\n")
	cmds := make([]string, 0, len(lines))
	last := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "\\") {
			last = last + line[0:len(line)-1]
		} else if len(last) > 0 {
			cmds = append(cmds, last+line)
			last = ""
		} else {
			cmds = append(cmds, line)
		}
	}
	if len(last) > 0 {
		// dangle.. itgnore for now
		cmds = append(cmds, last)
	}
	/*
		for i, ll := range cmds {
			log.Printf("LINE %d: %s", i+1, ll)
		}
	*/
	return cmds
}

func Move(s *Session, cli []string) error {
	name := cli[0]
	fargs := cli[1:]
	useGlob := false
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.BoolVar(&useGlob, "glob", false, "treat sources as globs")
	err := f.Parse(fargs)
	if err != nil {
		return err
	}
	args := f.Args()
	if len(args) < 2 {
		return fmt.Errorf("Expected at least 2 args")
	}

	dest, src := args[len(args)-1], args[:len(args)-1]
	if useGlob {
		sources := []string{}
		for _, val := range src {
			matches, err := filepath.Glob(val)
			if err != nil {
				return err
			}
			sources = append(sources, matches...)
		}
		if len(sources) == 0 {
			return fmt.Errorf("No matching files")
		}
		src = sources
	}

	if fileIsDirectory(dest) {
		for _, val := range src {
			base := filepath.Base(val)
			srcdest := filepath.Join(dest, base)
			os.Rename(val, srcdest)
			if err != nil {
				return fmt.Errorf("%s %s %s failed: %s",
					name, val, srcdest, err)
			}
		}
		return nil
	}

	// destination is not a directory
	if len(src) != 1 {
		return fmt.Errorf("Last arg is not a directory")
	}
	return os.Rename(src[0], dest)
}

func copyFile(src, dst string) error {
	log.Printf("---> Copying %s to %s", src, dst)
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func Copy(s *Session, cli []string) error {
	name := cli[0]
	fargs := cli[1:]
	useGlob := false
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.BoolVar(&useGlob, "glob", false, "treat sources as globs")
	err := f.Parse(fargs)
	if err != nil {
		return fmt.Errorf("%s: %s", name, err)
	}
	args := f.Args()
	if len(args) < 2 {
		return fmt.Errorf("%s: Expected at least 2 args", name)
	}

	dest, src := args[len(args)-1], args[:len(args)-1]
	if useGlob {
		sources := []string{}
		for _, val := range src {
			matches, err := filepath.Glob(val)
			if err != nil {
				return fmt.Errorf("%s: glob for %q failed: %s",
					name, val, err)
			}
			sources = append(sources, matches...)
		}
		if len(sources) == 0 {
			return fmt.Errorf("%s: No matching files for %v",
				name, src)
		}
		src = sources
	}

	if fileIsDirectory(dest) {
		for _, val := range src {
			base := filepath.Base(val)
			srcdest := filepath.Join(dest, base)
			copyFile(val, srcdest)
			if err != nil {
				return fmt.Errorf("%s %s %s failed: %s",
					name, val, srcdest, err)
			}
		}
		return nil
	}

	// destination is not a directory
	if len(src) != 1 {
		return fmt.Errorf("Last arg is not a directory")
	}
	return copyFile(src[0], dest)
}
