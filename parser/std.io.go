package parser

func makeIoLib() Module {
	m := Module{newObject()}
	m.addMember("log", Function{
		Params:   &Tuple{[]ExpressionType{Unknown{}}},
		Returned: Nil{},
	})
	return m
}
