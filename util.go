package gsh

import (
	"bufio"
	"io"

	"gopkg.in/pipe.v2"
)

// ForEachLine
//
// if we have args:
//    use args
// if we dont have args
//    run through each line
//
func ForEachLine(s *pipe.State, args []string, f func(line []byte) ([]byte, error)) error {
	for _, arg := range args {
		line, err := f([]byte(arg))
		if err != nil {
			return err
		}
		_, err = s.Stdout.Write(line)
		if err != nil {
			return err
		}
		s.Stdout.Write([]byte{'\n'})
	}

	// we are done
	if len(args) > 0 {
		return nil
	}

	scanner := bufio.NewScanner(s.Stdin)
	for scanner.Scan() {
		line, err := f(scanner.Bytes())
		if err != nil {
			return err
		}
		s.Stdout.Write(line)
		s.Stdout.Write([]byte{'\n'})
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}
