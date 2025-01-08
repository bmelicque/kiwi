package parser

import "maps"

func newGetter(t ExpressionType) Function {
	return Function{
		TypeParams: []Generic{},
		Params:     &Tuple{[]ExpressionType{}},
		Returned:   t,
	}
}

var domLib Module

func DomLib() Module {
	if len(domLib.Members) > 0 {
		return domLib
	}

	Event := Trait{
		Self:    Generic{Name: "Self"},
		Members: map[string]ExpressionType{},
	}
	EventHandler := Function{
		Params:   &Tuple{[]ExpressionType{Event}},
		Returned: Nil{},
	}
	EventTarget := Trait{}
	Node := Trait{
		Self:    Generic{Name: "Self"},
		Members: map[string]ExpressionType{},
	}
	Element := Trait{
		Self:    Generic{Name: "Self"},
		Members: map[string]ExpressionType{},
	}
	Document := Trait{
		Self:    Generic{Name: "Self"},
		Members: map[string]ExpressionType{},
	}

	Event.Members = map[string]ExpressionType{
		"bubbles":                  newGetter(Boolean{}),
		"cancelable":               newGetter(Boolean{}),
		"composed":                 newGetter(Boolean{}),
		"composedPath":             newGetter(List{Ref{EventTarget}}),
		"currentTarget":            newGetter(Ref{Generic{Name: "Self"}}),
		"defaultPrevented":         newGetter(Boolean{}),
		"eventPhase":               newGetter(Nil{}), // TODO: returns https://developer.mozilla.org/fr/docs/Web/API/Event/eventPhase
		"isTrusted":                newGetter(Boolean{}),
		"preventDefault":           newFunction(),
		"stopImmediatePropagation": newFunction(),
		"stopPropagation":          newFunction(),
		"target":                   newGetter(Ref{EventTarget}),
		"timeStamp":                newGetter(Number{}),
		"type":                     newGetter(String{}),
	}

	EventTarget.Members = map[string]ExpressionType{
		"addEventListener": Function{
			Params: &Tuple{[]ExpressionType{
				String{},
				EventHandler,
			}},
			Returned: Nil{},
		},
		"dispatchEvent": Function{
			Params: &Tuple{[]ExpressionType{
				Event,
			}},
			Returned: Nil{},
		},
		"removeEventListener": Function{
			Params: &Tuple{[]ExpressionType{
				String{},
				EventHandler,
			}},
			Returned: Nil{},
		},
	}

	Node.Members = map[string]ExpressionType{
		"appendChild": Function{
			Params:   &Tuple{[]ExpressionType{Ref{Node}}},
			Returned: Nil{},
		},
		"baseURI":    newGetter(String{}),
		"childNodes": newGetter(Ref{List{Ref{Node}}}),
		"cloneNode":  newGetter(Generic{Name: "Self"}),
		"compareDocumentPosition": Function{
			Params:   &Tuple{[]ExpressionType{Node}},
			Returned: Number{}, // TODO: return enum https://developer.mozilla.org/en-US/docs/Web/API/Node/compareDocumentPosition
		},
		"contains": Function{
			Params:   &Tuple{[]ExpressionType{Ref{Node}}},
			Returned: Boolean{},
		},
		"firstChild":    newGetter(Ref{Node}),
		"getRootNode":   newGetter(Ref{Node}), // TODO: options params https://developer.mozilla.org/en-US/docs/Web/API/Node/getRootNode
		"hasChildNodes": newGetter(Boolean{}),
		"insertBefore": Function{
			TypeParams: []Generic{{Name: "Inserted", Constraints: Ref{Node}}},
			Params: &Tuple{[]ExpressionType{
				Ref{Node},
				Generic{Name: "Inserted", Constraints: Ref{Node}},
			}},
			Returned: Generic{Name: "Inserted", Constraints: Ref{Node}},
		},
		"isDefaultNamespace": Function{
			Params:   &Tuple{[]ExpressionType{String{}}},
			Returned: Boolean{},
		},
		// TODO: isEqualNode? shouldn't '==' be enough?
		// TODO: isSameNode? shouldn't '&a == &b' be enough?
		"lastChild":       newGetter(Ref{Node}),
		"nextSibling":     newGetter(Ref{Node}),
		"nodeName":        newGetter(String{}),
		"nodeType":        newGetter(Number{}), // TODO: node type enum: https://developer.mozilla.org/fr/docs/Web/API/Node
		"normalize":       newFunction(),
		"ownerDocument":   newGetter(Ref{Document}),
		"parentNode":      newGetter(Ref{Node}),
		"parentElement":   newGetter(Nil{}), // TODO: &Element{}
		"previousSibling": newGetter(Ref{Node}),
		"removeChild": Function{
			TypeParams: []Generic{{Name: "Removed", Constraints: Ref{Node}}},
			Params:     &Tuple{[]ExpressionType{Generic{Name: "Removed", Constraints: Ref{Node}}}},
			Returned:   Generic{Name: "Removed", Constraints: Ref{Node}},
		},
		"replaceChild": Function{
			TypeParams: []Generic{{Name: "Replaced", Constraints: Ref{Node}}},
			Params: &Tuple{[]ExpressionType{
				Generic{Name: "Replaced", Constraints: Ref{Node}},
				Ref{Node},
			}},
			Returned: Generic{Name: "Replaced", Constraints: Ref{Node}},
		},

		// TODO: nodeValue => getter & setter
		// TODO: textContent => getter & setter
	}
	maps.Copy(Node.Members, EventTarget.Members)

	Document.Members = map[string]ExpressionType{
		"activeElement": newGetter(Nil{}), // TODO: returns &Element
	}

	Element.Members = map[string]ExpressionType{}
	maps.Copy(Element.Members, Node.Members)

	domLib.addMember("Event", Event)
	domLib.addMember("EventHandler", EventHandler)
	domLib.addMember("EventTarget", EventTarget)
	domLib.addMember("Node", Node)
	domLib.addMember("Element", Element)

	domLib.addMember("createElement", Function{
		Params:   &Tuple{[]ExpressionType{String{}}},
		Returned: Element,
	})

	return domLib
}
