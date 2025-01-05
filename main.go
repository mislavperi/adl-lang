package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/mislavperi/gem-lang/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	const GEM = `
    __________
   /          \
  /            \
 / /\        /\ \
/_/  \______/  \_\
\ \   /    \   / /
 \ \ /      \ / /
  \ /        \ /
   \          /
    \        /
     \      /
      \    /
       \  /
	\/	
`

	fmt.Printf(GEM)
	fmt.Printf("Hello %s! This is the Gem programming lanauage!\n", user.Username)
	fmt.Printf("Feel free to type in some commands\n")
	repl.Start(os.Stdin, os.Stdout)
}
