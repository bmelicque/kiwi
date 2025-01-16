package parser

func makeIoLib() Module {
	m := Module{newObject()}
	m.addMember("log", Function{
		Params:   &Tuple{[]ExpressionType{Invalid{}}},
		Returned: Void{},
	})
	return m
}
