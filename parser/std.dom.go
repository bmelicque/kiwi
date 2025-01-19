package parser

import (
	"fmt"
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
		From: "dom",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}
	domLib.addMember("Event", Type{Event})
	domLib.addMember("EventHandler", Function{
		Params:   &Tuple{[]ExpressionType{Event}},
		Returned: Void{},
	})
	domLib.addMember("EventTarget", Type{TypeAlias{
		Name: "EventTarget",
		From: "dom",
		Ref:  Trait{Members: map[string]ExpressionType{}},
	}})
	domLib.addMember("Node", Type{TypeAlias{
		Name: "Node",
		From: "dom",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}})
	Element := TypeAlias{
		Name: "Element",
		From: "dom",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}
	domLib.addMember("Element", Type{Element})
	HTMLBodyElement := TypeAlias{
		Name: "HTMLBodyElement",
		From: "dom",
		Ref:  newObject(),
	}
	domLib.addMember("HTMLBodyElement", Type{HTMLBodyElement})
	HTMLFrameElement := TypeAlias{
		Name: "HTMLFrameElement",
		From: "dom",
		Ref:  newObject(),
	}
	domLib.addMember("HTMLFrameElement", Type{HTMLFrameElement})
	domLib.addMember("CharacterData", Type{TypeAlias{
		Name: "CharacterData",
		From: "dom",
		Ref:  Trait{Self: Generic{Name: "Self"}, Members: map[string]ExpressionType{}},
	}})
	domLib.addMember("Text", Type{TypeAlias{
		Name:    "Text",
		From:    "dom",
		Ref:     Object{Members: []ObjectMember{{Name: "data", Type: String{}}}},
		Methods: map[string]ExpressionType{},
	}})
	Document := TypeAlias{
		Name:    "Document",
		From:    "dom",
		Ref:     newObject(),
		Methods: map[string]ExpressionType{},
	}
	domLib.addMember("Document", Type{Document})
	domLib.addMember("DocumentBody", Type{TypeAlias{
		Name: "DocumentBody",
		From: "dom",
		Ref: Sum{map[string]Tuple{
			"Body":  {[]ExpressionType{HTMLBodyElement}},
			"Frame": {[]ExpressionType{HTMLFrameElement}},
		}},
		Methods: map[string]ExpressionType{},
	}})
	HTMLError := TypeAlias{
		Name:    "HTMLError",
		From:    "dom",
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
	domLib.addMember("document", newGetter(Ref{Document}))

	return domLib
}

func buildEventTrait() {
	Event := getDomMember("Event")
	EventTarget := getDomMember("EventTarget")
	methods := Event.(TypeAlias).Ref.(Trait).Members

	methods["bubbles"] = newGetter(Boolean{})
	methods["cancelable"] = newGetter(Boolean{})
	methods["composed"] = newGetter(Boolean{})
	methods["composedPath"] = newGetter(List{Ref{EventTarget}})
	methods["currentTarget"] = newGetter(Ref{Generic{Name: "Self"}})
	methods["defaultPrevented"] = newGetter(Boolean{})
	methods["eventPhase"] = newGetter(Void{}) // TODO: returns https://developer.mozilla.org/fr/docs/Web/API/Event/eventPhase
	methods["isTrusted"] = newGetter(Boolean{})
	methods["preventDefault"] = newFunction()
	methods["stopImmediatePropagation"] = newFunction()
	methods["stopPropagation"] = newFunction()
	methods["target"] = newGetter(Ref{EventTarget})
	methods["timeStamp"] = newGetter(Number{})
	methods["type"] = newGetter(String{})
}

func buildEventTargetTrait() {
	EventTarget := getDomMember("EventTarget")
	EventHandler, _ := domLib.GetOwned("EventHandler")
	Event := getDomMember("Event")
	methods := EventTarget.(TypeAlias).Ref.(Trait).Members

	methods["addEventListener"] = Function{
		Params: &Tuple{[]ExpressionType{
			String{},
			EventHandler,
		}},
		Returned: Void{},
	}
	methods["dispatchEvent"] = Function{
		Params: &Tuple{[]ExpressionType{
			Event,
		}},
		Returned: Void{},
	}
	methods["removeEventListener"] = Function{
		Params: &Tuple{[]ExpressionType{
			String{},
			EventHandler,
		}},
		Returned: Void{},
	}
}

func buildNodeTrait() {
	Node := getDomMember("Node")
	Document := getDomMember("Document")
	Element := getDomMember("Element")
	EventTarget := getDomMember("EventTarget")
	methods := Node.(TypeAlias).Ref.(Trait).Members

	methods["appendChild"] = Function{
		Params:   &Tuple{[]ExpressionType{Ref{Node}}},
		Returned: Void{},
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
	methods["parentElement"] = newGetter(Ref{Element})
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

	maps.Copy(methods, EventTarget.(TypeAlias).Ref.(Trait).Members)
}

func buildDocumentType() {
	Document := getDomMember("Document")
	DocumentBody := getDomMember("DocumentBody")
	Element := getDomMember("Element")
	Node := getDomMember("Node")
	methods := Document.(TypeAlias).Methods

	methods["activeElement"] = newGetter(Ref{Element})
	methods["body"] = newGetter(makeOptionType(Ref{DocumentBody}))
	methods["setBody"] = newGetter(Function{
		Params:   &Tuple{[]ExpressionType{Ref{DocumentBody}}},
		Returned: Void{},
	})

	maps.Copy(methods, Node.(TypeAlias).Ref.(Trait).Members)
}

func buildElementTrait() {
	Element := getDomMember("Element")
	Node := getDomMember("Node")
	methods := Element.(TypeAlias).Ref.(Trait).Members

	// TODO: Element methods

	maps.Copy(methods, Node.(TypeAlias).Ref.(Trait).Members)
}

func buildCharacterDataTrait() {
	CharacterData := getDomMember("CharacterData")
	Node := getDomMember("Node")
	methods := CharacterData.(TypeAlias).Ref.(Trait).Members

	// TODO: CharacterData methods

	maps.Copy(methods, Node.(TypeAlias).Ref.(Trait).Members)
}

func buildTextType() {
	Text := getDomMember("Text")
	Element := getDomMember("Element")
	CharacterData := getDomMember("CharacterData")
	methods := Text.(TypeAlias).Methods

	methods["assignedSlot"] = Function{Params: &Tuple{}, Returned: Ref{Element}} // TODO: HTMLSlotElement
	methods["wholeText"] = Function{Params: &Tuple{}, Returned: Text}
	methods["splitText"] = Function{
		Params:   &Tuple{Elements: []ExpressionType{Number{}}},
		Returned: List{Text},
	}
	methods["cloneNode"] = newGetter(Text)

	maps.Copy(methods, CharacterData.(TypeAlias).Ref.(Trait).Members)
}

func getDomMember(name string) ExpressionType {
	t, ok := domLib.GetOwned(name)
	if !ok {
		panic(fmt.Sprintf("invalid member %v in dom lib", name))
	}
	return t.(Type).Value
}
