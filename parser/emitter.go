package parser

import "strings"

type EmitterFlag int

const (
	NoFlags   EmitterFlag = 0
	RangeFlag EmitterFlag = 1 << iota
)

type Emitter struct {
	depth   int
	flags   EmitterFlag
	builder strings.Builder
}

func MakeEmitter() *Emitter {
	return &Emitter{0, NoFlags, strings.Builder{}}
}

func (e *Emitter) AddFlag(flag EmitterFlag) {
	e.flags |= flag
}

func (e *Emitter) Write(str string) {
	e.builder.WriteString(str)
}

func (e *Emitter) Indent() {
	for i := 0; i < e.depth; i++ {
		e.builder.WriteString("    ")
	}
}

func (e Emitter) String() string {
	return e.builder.String()
}
