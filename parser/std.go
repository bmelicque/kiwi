package parser

func getLib(name string) (Module, bool) {
	switch name {
	case "dom":
		return makeDomLib(), true
	case "io":
		return makeIoLib(), true
	default:
		return Module{}, false
	}
}
