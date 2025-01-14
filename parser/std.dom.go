package parser

import (
	"maps"
)

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

	Event := TypeAlias{
		Name: "Event",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}
	domLib.addMember("Event", Type{Event})
	domLib.addMember("EventHandler", Function{
		Params:   &Tuple{[]ExpressionType{Event}},
		Returned: Nil{},
	})
	domLib.addMember("EventTarget", Type{TypeAlias{
		Name: "EventTarget",
		Ref:  Trait{Members: map[string]ExpressionType{}},
	}})
	domLib.addMember("Node", Type{TypeAlias{
		Name: "Node",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}})
	Element := TypeAlias{
		Name: "Element",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}
	domLib.addMember("Element", Type{Element})
	domLib.addMember("CharacterData", Type{TypeAlias{
		Name: "CharacterData",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}})
	domLib.addMember("Text", Type{TypeAlias{
		Name:    "Text",
		Ref:     Object{Members: []ObjectMember{{Name: "data", Type: String{}}}},
		Methods: map[string]ExpressionType{},
	}})
	domLib.addMember("Document", Type{TypeAlias{
		Name:    "Document",
		Ref:     newObject(),
		Methods: map[string]ExpressionType{},
	}})
	HTMLError := TypeAlias{
		Name:    "HTMLError",
		Ref:     Object{},
		Methods: map[string]ExpressionType{"error": newGetter(String{})},
	}
	domLib.addMember("HTMLError", Type{HTMLError})

	buildEventTrait()
	buildEventTargetTrait()
	buildNodeTrait()
	buildDocumentType()
	buildElementTrait()
	buildCharacterDataTrait()
	buildTextType()

	domLib.addMember("createElement", Function{
		Params:   &Tuple{[]ExpressionType{String{}}},
		Returned: makeResultType(Element, HTMLError),
	})

	return domLib
}

func buildEventTrait() {
	event, _ := domLib.GetOwned("Event")
	eventTarget, _ := domLib.GetOwned("EventTarget")
	EventTarget := eventTarget.(Type).Value
	methods := event.(Type).Value.(TypeAlias).Ref.(Trait).Members

	methods["bubbles"] = newGetter(Boolean{})
	methods["cancelable"] = newGetter(Boolean{})
	methods["composed"] = newGetter(Boolean{})
	methods["composedPath"] = newGetter(List{Ref{EventTarget}})
	methods["currentTarget"] = newGetter(Ref{Generic{Name: "Self"}})
	methods["defaultPrevented"] = newGetter(Boolean{})
	methods["eventPhase"] = newGetter(Nil{}) // TODO: returns https://developer.mozilla.org/fr/docs/Web/API/Event/eventPhase
	methods["isTrusted"] = newGetter(Boolean{})
	methods["preventDefault"] = newFunction()
	methods["stopImmediatePropagation"] = newFunction()
	methods["stopPropagation"] = newFunction()
	methods["target"] = newGetter(Ref{EventTarget})
	methods["timeStamp"] = newGetter(Number{})
	methods["type"] = newGetter(String{})
}

func buildEventTargetTrait() {
	eventTarget, _ := domLib.GetOwned("EventTarget")
	EventHandler, _ := domLib.GetOwned("EventHandler")
	event, _ := domLib.GetOwned("Event")
	Event := event.(Type).Value
	methods := eventTarget.(Type).Value.(TypeAlias).Ref.(Trait).Members

	methods["addEventListener"] = Function{
		Params: &Tuple{[]ExpressionType{
			String{},
			EventHandler,
		}},
		Returned: Nil{},
	}
	methods["dispatchEvent"] = Function{
		Params: &Tuple{[]ExpressionType{
			Event,
		}},
		Returned: Nil{},
	}
	methods["removeEventListener"] = Function{
		Params: &Tuple{[]ExpressionType{
			String{},
			EventHandler,
		}},
		Returned: Nil{},
	}
}

func buildNodeTrait() {
	node, _ := domLib.GetOwned("Node")
	Node := node.(Type).Value
	document, _ := domLib.GetOwned("Document")
	Document := document.(Type).Value
	eventTarget, _ := domLib.GetOwned("EventTarget")
	methods := Node.(TypeAlias).Ref.(Trait).Members

	methods["appendChild"] = Function{
		Params:   &Tuple{[]ExpressionType{Ref{Node}}},
		Returned: Nil{},
	}
	methods["baseURI"] = newGetter(String{})
	methods["childNodes"] = newGetter(Ref{List{Ref{Node}}})
	methods["cloneNode"] = newGetter(Generic{Name: "Self"})
	methods["compareDocumentPosition"] = Function{
		Params:   &Tuple{[]ExpressionType{Node}},
		Returned: Number{}, // TODO: return enum https://developer.mozilla.org/en-US/docs/Web/API/Node/compareDocumentPosition
	}
	methods["contains"] = Function{
		Params:   &Tuple{[]ExpressionType{Ref{Node}}},
		Returned: Boolean{},
	}
	methods["firstChild"] = newGetter(Ref{Node})
	methods["getRootNode"] = newGetter(Ref{Node}) // TODO: options params https://developer.mozilla.org/en-US/docs/Web/API/Node/getRootNode
	methods["hasChildNodes"] = newGetter(Boolean{})
	methods["insertBefore"] = Function{
		TypeParams: []Generic{{Name: "Inserted", Constraints: Ref{Node}}},
		Params: &Tuple{[]ExpressionType{
			Ref{Node},
			Generic{Name: "Inserted", Constraints: Ref{Node}},
		}},
		Returned: Generic{Name: "Inserted", Constraints: Ref{Node}},
	}
	methods["isDefaultNamespace"] = Function{
		Params:   &Tuple{[]ExpressionType{String{}}},
		Returned: Boolean{},
	}
	// TODO: isEqualNode? shouldn't '==' be enough?
	// TODO: isSameNode? shouldn't '&a == &b' be enough?
	methods["lastChild"] = newGetter(Ref{Node})
	methods["nextSibling"] = newGetter(Ref{Node})
	methods["nodeName"] = newGetter(String{})
	methods["nodeType"] = newGetter(Number{}) // TODO: node type enum: https://developer.mozilla.org/fr/docs/Web/API/Node
	methods["normalize"] = newFunction()
	methods["ownerDocument"] = newGetter(Ref{Document})
	methods["parentNode"] = newGetter(Ref{Node})
	methods["parentElement"] = newGetter(Nil{}) // TODO: &Element{}
	methods["previousSibling"] = newGetter(Ref{Node})
	methods["removeChild"] = Function{
		TypeParams: []Generic{{Name: "Removed", Constraints: Ref{Node}}},
		Params:     &Tuple{[]ExpressionType{Generic{Name: "Removed", Constraints: Ref{Node}}}},
		Returned:   Generic{Name: "Removed", Constraints: Ref{Node}},
	}
	methods["replaceChild"] = Function{
		TypeParams: []Generic{{Name: "Replaced", Constraints: Ref{Node}}},
		Params: &Tuple{[]ExpressionType{
			Generic{Name: "Replaced", Constraints: Ref{Node}},
			Ref{Node},
		}},
		Returned: Generic{Name: "Replaced", Constraints: Ref{Node}},
	}
	// TODO: nodeValue => getter & setter
	// TODO: textContent => getter & setter

	maps.Copy(methods, eventTarget.(Type).Value.(TypeAlias).Ref.(Trait).Members)
}

func buildDocumentType() {
	document, _ := domLib.GetOwned("Document")
	Document := document.(Type).Value
	node, _ := domLib.GetOwned("Node")
	methods := Document.(TypeAlias).Methods

	methods["activeElement"] = newGetter(Nil{}) // TODO: returns &Element

	maps.Copy(methods, node.(Type).Value.(TypeAlias).Ref.(Trait).Members)
}

func buildElementTrait() {
	element, _ := domLib.GetOwned("Element")
	Element := element.(Type).Value
	node, _ := domLib.GetOwned("Node")
	methods := Element.(TypeAlias).Ref.(Trait).Members

	// TODO: Element methods

	maps.Copy(methods, node.(Type).Value.(TypeAlias).Ref.(Trait).Members)
}

func buildCharacterDataTrait() {
	characterData, _ := domLib.GetOwned("CharacterData")
	CharacterData := characterData.(Type).Value
	node, _ := domLib.GetOwned("Node")
	methods := CharacterData.(TypeAlias).Ref.(Trait).Members

	// TODO: CharacterData methods

	maps.Copy(methods, node.(Type).Value.(TypeAlias).Ref.(Trait).Members)
}

func buildTextType() {
	text, _ := domLib.GetOwned("Text")
	Text := text.(Type).Value
	element, _ := domLib.GetOwned("Element")
	Element := element.(Type).Value
	characterData, _ := domLib.GetOwned("CharacterData")
	methods := Text.(TypeAlias).Methods

	methods["assignedSlot"] = Function{Params: &Tuple{}, Returned: Ref{Element}} // TODO: HTMLSlotElement
	methods["wholeText"] = Function{Params: &Tuple{}, Returned: Text}
	methods["splitText"] = Function{
		Params:   &Tuple{Elements: []ExpressionType{Number{}}},
		Returned: List{Text},
	}
	methods["cloneNode"] = newGetter(Text)

	maps.Copy(methods, characterData.(Type).Value.(TypeAlias).Ref.(Trait).Members)
}
