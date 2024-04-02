package main

import (
	"fmt"
	"log/slog"

	"github.com/moltencan/funnyfarm/cdb"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	c, err := cdb.NewDB("./test.db")
	if err != nil {
		panic(err)
	}

	fmt.Println(c)
}
