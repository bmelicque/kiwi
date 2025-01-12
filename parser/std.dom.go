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

	eventMembers := map[string]ExpressionType{}
	Event := TypeAlias{
		Name: "Event",
		Ref: Trait{
			Self:    Generic{Name: "Self"},
			Members: eventMembers,
		},
	}
	EventHandler := Function{
		Params:   &Tuple{[]ExpressionType{Event}},
		Returned: Nil{},
	}
	eventTargetMembers := map[string]ExpressionType{}
	EventTarget := TypeAlias{
		Name: "EventTarget",
		Ref:  Trait{Members: eventTargetMembers},
	}
	nodeMembers := map[string]ExpressionType{}
	Node := TypeAlias{
		Name: "Node",
		Ref: Trait{
			Self:    Generic{Name: "Self"},
			Members: nodeMembers,
		},
	}
	elementMembers := map[string]ExpressionType{}
	Element := TypeAlias{
		Name: "Element",
		Ref: Trait{
			Self:    Generic{Name: "Self"},
			Members: elementMembers,
		},
	}
	documentMembers := map[string]ExpressionType{}
	Document := TypeAlias{
		Name: "Document",
		Ref: Trait{
			Self:    Generic{Name: "Self"},
			Members: documentMembers,
		},
	}

	maps.Copy(eventMembers, map[string]ExpressionType{
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
	})

	maps.Copy(eventTargetMembers, map[string]ExpressionType{
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
	})

	maps.Copy(nodeMembers, map[string]ExpressionType{
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
	})
	maps.Copy(nodeMembers, eventTargetMembers)

	documentMembers["activeElement"] = newGetter(Nil{}) // TODO: returns &Element
	maps.Copy(documentMembers, nodeMembers)

	maps.Copy(elementMembers, nodeMembers)

	domLib.addMember("Event", Event)
	domLib.addMember("EventHandler", EventHandler)
	domLib.addMember("EventTarget", EventTarget)
	domLib.addMember("Node", Node)
	domLib.addMember("Element", Element)

	domLib.addMember("createElement", Function{
		Params:   &Tuple{[]ExpressionType{String{}}},
		Returned: Element, // TODO: this should be a Result
	})

	return domLib
}
