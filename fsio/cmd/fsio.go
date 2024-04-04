package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moltencan/funnyfarm/fsio"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal("push|pull")
	}

	dir := "./test"

	switch os.Args[1] {
	case "pull":
		fmt.Println("waiting for messages")
		err := fsio.Pull(dir, func(data []byte) error {
			fmt.Println(string(data))
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

	case "push":
		if err := fsio.Push(dir, []byte("hello, world")); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("invalid command")
	}
}
