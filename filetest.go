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
	_, err := exec.LookPath(path)
	return err == nil
}

func fileIsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileIsRegular(fname string) bool {
	info, err := os.Stat(fname)
	return err == nil && info.Mode().IsRegular()
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


func (s *Session) Test(str string) bool {
	// special replacement of environment variables.
	// in regular case we just ${foo} --> bar
	// in this case they are quoted ${foo} --> "bar"
	//  since they need to be golang proper values
	str = os.Expand(str, func(key string) string {
		return fmt.Sprintf("%q", s.Env[key])
	})

	t := template.New("gsh.test").Funcs(fmap)
	src := fmt.Sprintf("{{ if (%s) }}1{{ else }}0{{ end }}", str)
	t, err := t.Parse(src)
	if err != nil {
		s.SetError(fmt.Errorf("Unable to parse %q: %s", src, err))
		return false
	}
	out := bytes.Buffer{}
	err = t.Execute(&out, nil)
	if err != nil {
		s.SetError(err)
		return false
	}
	result := out.String()
	switch result {
	case "0":
		return false
	case "1":
		return true
	default:
		s.SetError(fmt.Errorf("Unknown value from test: %q", result))
		return false
	}
}
