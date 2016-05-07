/*
Package sherlock a simple little package designed to help tidy up go code by reducing
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
	// ErrInspect is returned if Inspect is called and the final argument
	// is not an error type (or nil).
	ErrInspect = errors.New("improper use of Inspect function")
)

// Sherlock checks errors for you
type Sherlock struct {
	notebook string
	action   func(error)
}

type failure struct {
	err   error
	stack []byte
}

// Notebook can be called to set a file location where Sherlock should leave
// his notes after an investigation. If an error occurs whilst trying to use
// this file, Sherlock will revert to creating a temporary file for it.
func (s *Sherlock) Notebook(path string) {
	s.notebook = path
}

// Action sets an action for Sherlock to perform after concluding an
// investion if something went wrong.
func (s *Sherlock) Action(fn func(err error)) {
	s.action = fn
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
func (s *Sherlock) Investigation() {
	r := recover()
	if r != nil {
		fail, ok := r.(*failure)
		if !ok {
			panic(r)
		}
		s.writeCaseFiles(fail)
		s.action(fail.err)
	}
}

func (s *Sherlock) writeCaseFiles(fail *failure) {
	var err error
	var notebook *os.File
	if s.notebook != "" {
		err = os.Remove(s.notebook)
		if err == nil {
			notebook, err = os.Create(s.notebook)
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
