package gsh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"
)

// BUG: does not use local session path

func commandExists(path string) bool {
	return exec.LookPath(path)
}

func fileIsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileIsRegular(fname string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsRegular()
}

func fileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
}

var fmap = template.FuncMap{
	"fileIsRegular":   fileIsRegular,
	"fileIsDirectory": fileIsDirectory,
	"fileExists":      fileExists,
	"commandExists":   commandExists,
	"basename":        path.Base,
}

func (s *Session) Test(s string) (bool, error) {
	// special replacement of environment variables.
	// in regular case we just ${foo} --> bar
	// in this case they are quoted ${foo} --> "bar"
	//  since they need to be golang proper values
	s = os.Expand(s, (func(key string) string {
		return fmt.Printf("%q", s.Env[key])
	}))

	t := template.New("nothing").Funcs(fmap)
	t, err := t.Parse(fmt.Sprintf("{{ if %s }}1{{ else }}0{{ end }}", s))
	if err != nil {
		return false, err
	}
	out := bytes.Buffer{}
	err = t.Execute(out, nil)
	if err != nil {
		return false, err
	}
	result := out.String()
	switch result {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("Unknown value from test: %q", result)
	}
}
