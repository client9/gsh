package main

import (
	"flag"
	"log"
	"os"

	"github.com/client9/gsh"
	"gopkg.in/pipe.v2"
)


func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("OOPS no command")
	}

	cmd, ok := gsh.CmdMap[args[0]]
	if !ok {
		log.Fatalf("Unknown command %s", args[0])
	}

	p := cmd(args[1:]...)

	s := pipe.NewState(os.Stdout, os.Stderr)
	s.Stdin = os.Stdin
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
