package argstruct

type ArgStructable interface {
	Run(*ArgStruct)
}

func Run(a ArgStructable) {

}

type ArgStruct struct {
	as ArgStructable
}

func (x *ArgStruct) Run() {
	x.as.Run(x)
}
