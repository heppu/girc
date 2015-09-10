package main

import (
	"bufio"
	"fmt"
	. "github.com/heppu/girc"
	"os"
)

func main() {
	// Config
	// s := "irc.atw-inter.net:6667"
	s := "irc.ca.ircnet.net:6667"
	n := "heppu_girc"
	r := "Heppu Girc"

	// Client
	c, err := NewClient(s, n, r)

	if err != nil {
		panic(err)
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		switch text {
		case "c\n":
			c.Connect()
		case "j\n":
			c.Join()
		case "h\n":
			c.Send()
		default:
			c.Quit()
			return
		}
	}
	fmt.Println("Done")
}
