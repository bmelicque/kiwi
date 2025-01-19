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
func createStdName(rootDir string) string {
	for {
		name := createRandomString(8) + ".js"
		if fileDoesNotExists(filepath.Join(rootDir, createRandomString(8))) {
			return name
		}
	}
}

// returns the path to the .js file containing std lib
func EmitStd(rootDir, outDir string) string {
	name := createStdName(rootDir)
	f, err := os.Create(filepath.Join(outDir, name))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// sum types
	f.WriteString(`export function Sum(t,v){this.tag=t;this.value=v}
export class Pointer{constructor(c,n){this.c=c;this.n=n}get(){return this.c?.[this.n]??this.n}set(v){this.c?(this.c[this.n]=v):(this.n=v)}}
export class NodePointer{constructor(v){this._=v}get(){return this._}set(v){this._.parentNode?.replaceChild(this._,v);this._=v}}
export let n=NodePointer,wrapNodeMethod=(o,m,r,f=(...a)=>o[m].apply(o,a.map(a=>a instanceof n?a.get():a)))=>(typeof o[m]!="function"?r?()=>new n(o[m]):()=>o[m]:r?(...a)=>new n(f(...a)):f),bind=(o,m)=>(o[m].bind(o)),equals=(a,b,t=typeof a)=>(t==typeof b&&(t!="object"||a==null||b==null?a==b:a.constructor==b.constructor&&(a instanceof n?a.get()==b.get():!(Array.isArray(a)&&a.length-b.length)&&!Object.keys(a).find(k=>!equals(a[k],b[k]))))),createElement=s=>{let[a,t,i,c]=s.match(/^(\w[\w\-_]*)?(?:#(\w[\w\-_]*))?((?:\.\w[\w\-_]*)*)$/);if(!a)throw new Error("Invalid selector");let e=document.createElement(t||"div");if(i)e.id=i.slice(1);if(c)e.classList.add(...c.split(".").slice(1));return e}
`)
	return name
}

type StandardFlags = uint

const (
	NoFlag StandardFlags = 0

	SumFlag StandardFlags = 1 << (iota - 1)
	PointerFlag
	NodePointerFlag

	DeepEqualFlag
	WrapNodeMethodFlag
	BindFlag
	CreateElementFlag
)

type stdEmitter struct {
	flags StandardFlags
}

func (e *stdEmitter) addFlag(flag StandardFlags) {
	e.flags |= flag
}

func (e *stdEmitter) hasFlag(flag StandardFlags) bool {
	return (e.flags & flag) != NoFlag
}
