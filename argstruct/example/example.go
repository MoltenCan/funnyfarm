package main

import (
	"fmt"

	"github.com/moltencan/funnyfarm/argstruct"
)

type MyApp struct {
	S          string `argstruct:"help=string to print"`
	Bob        string `argstruct:"help=bob lives here,andShort"`
	YouMomma   string `argstruct:"help=yomomma,pos=1"`
	Two        string `argstruct:"help=aaa,pos=2"`
	Foo        string `argstruct:"help=Foo Fighters,default=bar"`
	Zee        int    `argstruct:"help=a number,required,andShort"`
	SomeOption bool   `argstruct:"help=an option,andshort"`
}

func (x *MyApp) Run(as *argstruct.ArgStruct) error {
	fmt.Printf("%+v\n", x)
	return nil
}

func main() {
	argstruct.Run(&MyApp{})
}
