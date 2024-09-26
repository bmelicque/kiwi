package checker

func makeOptionType(t ExpressionType) TypeAlias {
	return TypeAlias{
		Name:   "Option",
		Params: []Generic{{Name: "Type", Value: t}},
		Ref: Sum{map[string]ExpressionType{
			"Some": Type{Generic{Name: "Type", Value: t}},
			"None": nil,
		}},
	}
}

var optionType = makeOptionType(nil)

var std = Scope{
	variables: map[string]*Variable{
		"List": {
			typing: Type{TypeAlias{
				Name:   "List",
				Params: []Generic{{Name: "Type"}},
				Ref:    List{Generic{Name: "Type"}},
			}},
		},
		"Option": {
			typing: Type{optionType},
		},
		"Result": {
			typing: Type{TypeAlias{
				Name:   "Result",
				Params: []Generic{{Name: "Ok"}, {Name: "Err"}},
				Ref: Sum{map[string]ExpressionType{
					"Ok":  Type{Generic{Name: "Ok"}},
					"Err": Type{Generic{Name: "Err"}},
				}},
			}},
		},
	},
}
