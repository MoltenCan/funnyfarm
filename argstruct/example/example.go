package main

import (
	"fmt"

	"github.com/moltencan/funnyfarm/argstruct"
)

type MyApp struct {
	S        string `argstruct:"help=string to print"`
	Bob      string `argstruct:"help=bob lives here,andShort"`
	YouMomma string `argstruct:"help=yomomma,pos=1"`
	Foo      string `argstruct:"help=Foo Fighters,default=bar"`
	Zee      int    `argstruct:"help=a number,required,andShort"`
}

func (x *MyApp) Run(as *argstruct.ArgStruct) error {
	println("Hello, world!")
	fmt.Printf("%+v", x)
	return nil
}

func main() {
	argstruct.Run(&MyApp{})
}
