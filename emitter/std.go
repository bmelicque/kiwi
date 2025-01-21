package emitter

import (
	"errors"
	"math/rand"
	"os"
	"path/filepath"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func createRandomString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = chars[rand.Intn(len(chars))]
	}
	return string(bytes)
}
func fileDoesNotExists(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, os.ErrNotExist)
}
func CreateStdName(rootDir string) string {
	for {
		name := createRandomString(8) + ".js"
		if fileDoesNotExists(filepath.Join(rootDir, createRandomString(8))) {
			return name
		}
	}
}

// returns the path to the .js file containing std lib
func EmitStd(filePath string, flags StandardFlags) {
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if hasFlag(flags, SumFlag|OptionFlag) {
		f.WriteString("export function Sum(t,v){this.tag=t;this.value=v}\n")
	}
	if hasFlag(flags, OptionFlag) {
		f.WriteString("export class Option extends Sum{}\n")
	}
	if hasFlag(flags, PointerFlag) {
		f.WriteString("export class Pointer{constructor(c,n){this.c=c;this.n=n}get(){return this.c?.[this.n]??this.n}set(v){this.c?(this.c[this.n]=v):(this.n=v)}}\n")
	}
	if hasFlag(flags, NodePointerFlag|DeepEqualFlag|WrapNodeMethodFlag) {
		f.WriteString("export class NodePointer{constructor(v){this._=v}get(){return this._}set(v){this._.parentNode?.replaceChild(this._,v);this._=v}}\n")
	}
	if hasFlag(flags, DeepEqualFlag) {
		f.WriteString(`export let equals=(a,b,t=typeof a)=>(t==typeof b&&(t!="object"||a==null||b==null?a==b:a.constructor==b.constructor&&(a instanceof NodePointer?a.get()==b.get():!(Array.isArray(a)&&a.length-b.length)&&!Object.keys(a).find(k=>!equals(a[k],b[k])))))` + "\n")
	}
	if hasFlag(flags, WrapNodeMethodFlag) {
		f.WriteString(`export let wrapNodeMethod=(o,m,r,n=NodePointer,f=(...a)=>o[m].apply(o,a.map(a=>a instanceof n?a.get():a)))=>(typeof o[m]!="function"?r?()=>new n(o[m]):()=>o[m]:r?(...a)=>new n(f(...a)):f)` + "\n")
	}
	if hasFlag(flags, BindFlag) {
		f.WriteString("export let bind=(o,m)=>(o[m].bind(o))\n")
	}
	if hasFlag(flags, DocumentFlag) {
		f.WriteString("export let getDocument=()=>new NodePointer(document)\n")
	}
	if hasFlag(flags, DocumentBodyFlag) {
		f.WriteString(`export class DocumentBody extends Sum{}` + "\n")
	}
	if hasFlag(flags, DocumentGetBodyFlag) {
		f.WriteString(`export let getDocumentBody=d=>()=>(d instanceof NodePointer&&(d=d.get()),d.body&&new DocumentBody(d.body instanceof HTMLBodyElement?"Body":"Frame",d.body))` + "\n")
	}
	if hasFlag(flags, DocumentSetBodyFlag) {
		f.WriteString("export let setDocumentBody=d=>(d instanceof NodePointer&&(d=d.get()),b=>d.body=b.value)\n")
	}
	if hasFlag(flags, CreateElementFlag) {
		f.WriteString(`export let createElement=s=>{let[a,t,i,c]=s.match(/^(\w[\w\-_]*)?(?:#(\w[\w\-_]*))?((?:\.\w[\w\-_]*)*)$/);if(!a)throw new Error("Invalid selector");let e=document.createElement(t||"div");if(i)e.id=i.slice(1);if(c)e.classList.add(...c.split(".").slice(1));return e}` + "\n")
	}
}

type StandardFlags = uint

const (
	NoFlag StandardFlags = 0

	SumFlag StandardFlags = 1 << (iota - 1)
	OptionFlag
	PointerFlag
	NodePointerFlag

	DeepEqualFlag
	WrapNodeMethodFlag
	BindFlag
	DocumentFlag
	DocumentBodyFlag
	DocumentGetBodyFlag
	DocumentSetBodyFlag
	CreateElementFlag
)

type stdEmitter struct {
	flags StandardFlags
}

func (e *stdEmitter) addFlag(flag StandardFlags) {
	e.flags |= flag
}

func hasFlag(flags, flag StandardFlags) bool {
	return (flags & flag) != NoFlag
}
