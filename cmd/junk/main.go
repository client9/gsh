package main

import (
	"fmt"

	"gopkg.in/pipe.v2"
	"github.com/client9/gsh"
)

func main() {
	p := gsh.Run("cat /dev/urandom | head -c 10 | base64")
	p = gsh.Run("echo 'Mon Apr 25 17:12:38 PDT 2016' | strptime -in 'Mon Jan _2 15:04:05 MST 2006'")
	/*
	p := pipe.Line(
		gsh.Run("cat /dev/urandom"),
		gsh.Run("head -c 10"),
		gsh.Run("base64"),
	)

	gsh.Cat("/dev/urandom"),
		//pipe.ReadFile("/dev/urandom"),
		gsh.Head("-c", "10"),
		gsh.Base64(),
	)
	p = pipe.Line(
		pipe.Println("Mon Apr 25 17:12:38 PDT 2016"),
		gsh.ParseTime("-in", "Mon Jan _2 15:04:05 MST 2006"),
	)
*/
	//p = pipe.Line(
	//	gsh.GitLastModified("./README.mdx"),
	//)
	out, err := pipe.CombinedOutput(p)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	fmt.Printf("out: %s\n", out)
}
