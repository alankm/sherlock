/*
Sherlock a simple little package designed to help tidy up go code by reducing
the substantial number of if err != nil checks usually performed.
*/
package sherlock

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
)

const ()

var (
	sherlock struct {
		notebook string
	}

	ErrInspect = errors.New("improper use of Inspect function")
)

type failure struct {
	err   error
	stack []byte
}

// Notebook can be called to set a file location where Sherlock should leave
// his notes after an investigation. If an error occurs whilst trying to use
// this file, Sherlock will revert to creating a temporary file for it.
func Notebook(path string) {
	sherlock.notebook = path
}

// Assert is used to ensure that things are operating as expected. If the
// statement proves to be false, then Assert throws a panic with the given err
// as its argument.
func Assert(statement bool, err error) {
	if statement == false {
		panic(&failure{err, debug.Stack()})
	}
}

// Inspect should be used with a function that can return an error. The final
// argument is assumed to be type error or nil. If it is an error, Inspect
// throws a panic with the given error as its argument.
func Inspect(vals ...interface{}) {
	x := vals[len(vals)-1]
	if x != nil {
		err, ok := x.(error)
		Assert(ok, ErrInspect)
		panic(&failure{err, debug.Stack()})
	}
}

// Investigation should be deferred before any
func Investigation() {
	r := recover()
	if r != nil {
		fail, ok := r.(*failure)
		if !ok {
			return
		}
		writeCaseFiles(fail)
	}
}

func writeCaseFiles(fail *failure) {
	var err error
	var notebook *os.File
	if sherlock.notebook != "" {
		err = os.Remove(sherlock.notebook)
		if err == nil {
			notebook, err = os.Create(sherlock.notebook)
		}
	}
	if notebook == nil {
		notebook, err = ioutil.TempFile("", "Sherlock-")
		if err != nil {
			panic(err)
		}
	}
	defer notebook.Close()

	fmt.Fprintf(notebook, "FAILURE: %v\n", fail.err.Error())
	fmt.Fprintf(notebook, "STACK TRACE:\n%v\n", string(fail.stack))
}
