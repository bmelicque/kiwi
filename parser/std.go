package parser

func getLib(name string) (Module, bool) {
	switch name {
	case "dom":
		return makeDomLib(), true
	default:
		return Module{}, false
	}
}
