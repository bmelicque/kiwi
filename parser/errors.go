package parser

import "fmt"

type ErrorKind = uint

const (
	NoError ErrorKind = iota

	CannotResolvePath // node should be *Literal containing path

	TokenExpected
	LeftBraceExpected
	RightBraceExpected
	RightBracketExpected

	ExpressionExpected
	UnexpectedExpression
	IntegerExpected
	StringLiteralExpected
	IdentifierExpected
	TypeIdentifierExpected
	ValueIdentifierExpected
	TypeParamsExpected
	RangeExpected
	FieldExpected
	FieldKeyExpected
	FunctionExpressionExpected
	CallExpressionExpected
	ParameterExpected
	ReceiverExpected

	IllegalBreak
	IllegalContinue
	IllegalReturn
	IllegalThrow
	IllegalResult

	ReservedName
	DuplicateIdentifier // [name]

	InvalidPattern
	InvalidTypeForPattern
	TooManyElements // [expected count, received count]
	MissingElements // [expected count, received count]
	MissingDefault
	UnreachableCode
	CatchallNotLast
	NotExhaustive

	InvalidAssignmentToEntry
	NonConstantTypeDeclaration
	NonConstantMethodDeclaration
	AssignmentToConstant

	TypeExpected
	ValueExpected
	BooleanExpected
	TypeOrBoolExpected
	NumberExpected // [got]
	IndexExpected
	ConcatenableExpected
	IterableExpected
	FunctionExpected
	PromiseExpected
	ResultExpected
	RefExpected
	ObjectTypeExpected
	FunctionTypeExpected

	ResultDeclaration
	VoidAssignment
	OrphanMethod
	UnneededCatch
	UnneededAsync
	UnusedVariable // [variable name]
	CannotFind

	OutOfRange
	MissingTypeArgs
	UnexpectedTypeArgs
	CannotAssignType // [expected type, received type]
	NotSubscriptable
	NotInstanceable
	Unmatchable
	NotReferenceable
	MismatchedTypes
	PropertyDoesNotExist
	MultipleEmbeddedProperties // [name]
	NotInModule                // [expected variable name]
	ModuleWrite
	PrivateProperty // [property name, path to origin file]
	PublicDeclaration
	TypeDoesNotImplement
	MissingKeys
	MissingConstructor
)

type ParserError struct {
	Node        Node
	Kind        ErrorKind
	Complements [2]interface{}
}

func (p ParserError) Text() string {
	switch p.Kind {
	case CannotResolvePath:
		return fmt.Sprintf("Cannot resolve path to %v", p.Node.(*Literal).Text())

	case TokenExpected:
		expected := p.Complements[0].(Token).Text()
		return fmt.Sprintf("'%v' expected", expected)
	case LeftBraceExpected:
		return "'{' expected"
	case RightBraceExpected:
		return "'}' expected"
	case RightBracketExpected:
		return "']' expected"

	case ExpressionExpected:
		return "Expression expected"
	case UnexpectedExpression:
		return "No expression expected"
	case IntegerExpected:
		return "Integer expected"
	case StringLiteralExpected:
		return "String literal expected"
	case IdentifierExpected:
		return "Identifier expected"
	case TypeIdentifierExpected:
		return "Type identifier expected"
	case ValueIdentifierExpected:
		return "Value identifier expected"
	case TypeParamsExpected:
		return "Type params expected"
	case RangeExpected:
		return "Range literal expression expected"
	case FieldExpected:
		return "Field expected"
	case FieldKeyExpected:
		return "Field key expected (identifier, literal or brackets)"
	case FunctionExpressionExpected:
		return "Function expression expected"
	case CallExpressionExpected:
		return "Call expression expected"
	case ParameterExpected:
		return "Parameter expected"
	case ReceiverExpected:
		return "Receiver param expected"

	case IllegalBreak:
		return "Cannot use 'break' keyword outside of a loop"
	case IllegalContinue:
		return "Cannot use 'continue' keyword outside of a loop"
	case IllegalReturn:
		return "Cannot use 'return' keyword outside of functions with explicit returns"
	case IllegalThrow:
		return "Cannot use 'throw' keyword outside of functions with explicit returns"
	case IllegalResult:
		return "Cannot use failable expressions outside of functions with explicit returns"

	case ReservedName:
		return fmt.Sprintf("'%v' is a reserved name", p.Complements[0])
	case DuplicateIdentifier:
		return fmt.Sprintf("Duplicate identifier '%v'", p.Complements[0])
	case InvalidPattern:
		return "Invalid pattern"
	case InvalidTypeForPattern:
		_ = p.Complements[0]
		assignedType := p.Complements[1].(ExpressionType).Text()
		return fmt.Sprintf("Cannot assign this value (%v) to that pattern", assignedType)
	case TooManyElements:
		a := p.Complements[0]
		b := p.Complements[1]
		return fmt.Sprintf("Got too many elements: expected %v, got %v", a, b)
	case MissingElements:
		a := p.Complements[0]
		b := p.Complements[1]
		return fmt.Sprintf("Got too few elements: expected %v, got %v", a, b)
	case MissingDefault:
		return "Private fields must have a default value"
	case UnreachableCode:
		return "Unreachable code detected"
	case CatchallNotLast:
		return "Catch-all case should be last"
	case NotExhaustive:
		return "Non-exhaustive match, consider adding a catch-all case"

	case InvalidAssignmentToEntry:
		return "Invalid assignment to entry; expected assignment to map entry"
	case NonConstantTypeDeclaration:
		return "Cannot declare mutable type, use '::' instead"
	case NonConstantMethodDeclaration:
		return "Cannot declare mutable methods, use '::' instead"
	case AssignmentToConstant:
		return "Cannot assign new value to constant"

	case TypeExpected:
		return "Type expected, got value"
	case ValueExpected:
		return "Value expected, got type"
	case BooleanExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("boolean expected, got %v", got)
	case TypeOrBoolExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Type or boolean expected, got %v", got)
	case NumberExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("number expected, got %v", got)
	case IndexExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("number or range expected, got %v", got)
	case ConcatenableExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Concatenable (string or list) expected, got %v", got)
	case IterableExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Iterable (list or slice) expected, got %v", got)
	case FunctionExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Function expected, got %v", got)
	case PromiseExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Promise expected, got %v", got)
	case ResultExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Result expected, got %v", got)
	case RefExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Reference expected, got %v", got)
	case ObjectTypeExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Object type expected, got %v", got)
	case FunctionTypeExpected:
		got := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Function type expected, got %v", got)

	case ResultDeclaration:
		return "Cannot declare a variable as a result type; consider using 'try' or 'catch'"
	case VoidAssignment:
		return "Cannot declare a variable as nil value; consider using the option type"
	case OrphanMethod:
		return "Methods have to be declared in the same file as the type they're attached to"
	case UnneededCatch:
		return "Unneeded catch (lhs is not a result type)"
	case UnneededAsync:
		return "Unneeded 'async' keyword"
	case UnusedVariable:
		return fmt.Sprintf("Unused variable '%v'", p.Complements[0])
	case CannotFind:
		return fmt.Sprintf("Cannot find name '%v'", p.Complements[0])

	case OutOfRange:
		return fmt.Sprintf("Index out of range: max %v, got %v", p.Complements[0], p.Complements[1])
	case MissingTypeArgs:
		return "Cannot fully determine type; probably missing some type arguments"
	case UnexpectedTypeArgs:
		return "No type arguments expected for this type"
	case CannotAssignType:
		t1, ok := p.Complements[0].(ExpressionType)
		var s1 string
		if ok {
			s1 = t1.Text()
		}
		t2, ok := p.Complements[1].(ExpressionType)
		var s2 string
		if ok {
			s2 = t2.Text()
		}
		return fmt.Sprintf("Cannot use value of type %v as type %v", s2, s1)
	case NotSubscriptable:
		t := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Type %v is not subscriptable", t)
	case NotInstanceable:
		t := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Type %v cannot be instanciated", t)
	case Unmatchable:
		t := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Cannot match against type %v", t)
	case NotReferenceable:
		return "Cannot reference such an expression"
	case MismatchedTypes:
		t1 := p.Complements[0].(ExpressionType).Text()
		t2 := p.Complements[1].(ExpressionType).Text()
		return fmt.Sprintf("Types %v and %v do not match", t1, t2)
	case PropertyDoesNotExist:
		name := p.Complements[0]
		parent := p.Complements[1].(ExpressionType).Text()
		return fmt.Sprintf("Property '%v' does not exist on type %v", name, parent)
	case MultipleEmbeddedProperties:
		name := p.Complements[0]
		// parent := p.Complements[1].(ExpressionType).Text()
		return fmt.Sprintf("Found several embedded properties with name '%v', consider fully qualifying the property", name)
	case NotInModule:
		variableName := p.Complements[0]
		return fmt.Sprintf("Variable '%v' does not exist in this module", variableName)
	case ModuleWrite:
		return "Cannot write module property"
	case PrivateProperty:
		key := p.Complements[0]
		path := p.Complements[1]
		return fmt.Sprintf("Property '%v' is private and cannot be used outside of its declaration file (%v)", key, path)
	case PublicDeclaration:
		return "Cannot declare public variables at top-level, consider making it private and defining a getter/setter"
	case TypeDoesNotImplement:
		name := p.Complements[0].(ExpressionType).Text()
		return fmt.Sprintf("Type %v does not implement this trait", name)
	case MissingKeys:
		return fmt.Sprintf("Missing key(s) %v", p.Complements[0])
	case MissingConstructor:
		return fmt.Sprintf("Missing constructor '%v'", p.Complements[0])

	default:
		panic("Error type not implemented")
	}
}
