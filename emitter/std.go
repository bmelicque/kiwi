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
	f.WriteString("export function Sum(t,v){this.tag=t;if(arguments.length>1)this.value=v}\n")

	// method for handling dom Node getter properties
	f.WriteString("export let callNodeGetter=(o,m)=>typeof o[m]==\"function\"?o[m]():o[m]\n")
	f.WriteString("export let bindNodeGetter=(o,m)=>typeof o[m]==\"function\"?o[m].bind(o):()=>o[m]\n")

	f.WriteString("export let bind=(o,m)=>o[m].bind(o)\n")

	// creating "pointers"
	f.WriteString("export class Pointer{constructor(c,n){this.c=c;this.n=n}get(){return this.c[this.n]}set(v){this.c[this.n]=v}}\n")

	return name
}
