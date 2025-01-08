package main

import "github.com/moltencan/funnyfarm/argstruct"

type Example2 struct {
	S    string `argstruct:"help=string to print"`
	Bob  bool   `argstruct:"group=group1,help=do a bob,andShort"`
	Fred bool   `argstruct:"group=group1,help=do a fred,andShort"`
}

func (x *Example2) Version() string {
	return "1.0.0"
}
func (x *Example2) Run(*argstruct.ArgStruct) error {
	switch {
	case x.Bob:
		println("bob")
	case x.Fred:
		println("fred")
	}
	return nil
}

func main() {
	argstruct.Run(&Example2{})
}
