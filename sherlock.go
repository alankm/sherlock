/*
Package sherlock a simple little package designed to help tidy up go code by reducing
the substantial number of if err != nil checks usually performed.
*/
package sherlock

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
)

// Sherlock checks errors for you
type Sherlock struct {
	notebook   string
	action     func(bool, error)
	registered map[error]bool
	mappings   map[error]error
	substrings map[string]error
	standard   error
}

// Standard allows you to define a default-case error for Sherlock to throw as a
// result of an Assert or a Try when the received error is not a registered
// error. The default behaviour is to just pass it straight through, but this
// makes it possible to pass a generic error such as an errInternal to represent
// an unexpected edge case that needs to be reported as a bug.
func (s *Sherlock) Standard(err error) {
	s.standard = err
}

// Register adds the given error to the set of errors that should not be
// replaced by the default error.
func (s *Sherlock) Register(err error) {
	s.registered[err] = true
}

// Map one error to another for Sherlock to translate. When Sherlock receives a
// registered error via Assert or Try, instead of panicking with the received
// error it will use the mapped error.
func (s *Sherlock) Map(input, output error) {
	s.mappings[input] = output
}

// MapPrefix a mapping from one substring to an error for Sherlock to translate.
// When Sherlock receives an error via Assert or Try and cannot find a
// Registered error or mapping, it will then resort to comparing all registered
// prefix strings against the Error() of the received error. If the received
// error contains a registered string as a prefix substring then it will use the
// error registered with this function.
func (s *Sherlock) MapPrefix(input string, output error) {
	s.substrings[input] = output
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
func (s *Sherlock) Action(fn func(detected bool, err error)) {
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

// Try should be used with a function that can return an error. The final
// argument is assumed to be type error or nil. If it is an error, Try throws a
// panic with the given error as its argument.
func Try(vals ...interface{}) {
	x := vals[len(vals)-1]
	if x != nil {
		err, ok := x.(error)
		if !ok {
			return
		}
		panic(&failure{err, debug.Stack()})
	}
}

func (s *Sherlock) error(err error) error {
	// check if the error is registered
	_, ok := s.registered[err]
	if ok {
		return err
	}
	// check if map contains err
	val, ok := s.mappings[err]
	if ok {
		return val
	}
	// check if err matches any known substring prefixes
	e := err.Error()
	for key, val := range s.substrings {
		if strings.HasPrefix(e, key) {
			return val
		}
	}
	// return default error if there is one
	if s.standard != nil {
		return s.standard
	}
	// return received error
	return err
}

// Investigation should be deferred before any
func (s *Sherlock) Investigation() {
	r := recover()
	if r != nil {
		fail, ok := r.(*failure)
		if !ok {
			fmt.Println(string(debug.Stack()))
			panic(r)
		}
		s.writeCaseFiles(fail)
	}
}

// Catch should be deferred and can be used within a closure to change the
// return value of an error.
func (s *Sherlock) Catch(err *error) {
	r := recover()
	if r == nil {
		return
	}
	x, ok := r.(error)
	if ok {
		*err = s.error(x)
		return
	}
	y, ok := r.(failure)
	if ok {
		*err = s.error(y.err)
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
