/*
Package sherlock helps tidy up go code by reducing the substantial number of
"if err != nil" checks usually performed.
*/
package sherlock

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
)

var (
	ErrUnexpected = errors.New("sherlock found an unexpected error")

	errUnrecoverable = errors.New("an unrecoverable bug occurred")

	packages map[string]*sherlock
)

type sherlock struct {
	errors         map[error]bool
	mappings       map[error]error
	regexps        map[string]bool
	regexpMappings map[string]error
}

func newSherlock() *sherlock {
	s := new(sherlock)
	s.errors = make(map[error]bool)
	s.mappings = make(map[error]error)
	s.regexps = make(map[string]bool)
	s.regexpMappings = make(map[string]error)
	return s
}

func Register(err error) {
	s := handler()
	s.errors[err] = true
}

func RegisterRegex(regex string) {
	s := handler()
	s.regexps[regex] = true
}

func RegisterMapping(x, y error) {
	s := handler()
	s.mappings[x] = y
}

func RegisterRegexMapping(x string, y error) {
	s := handler()
	s.regexpMappings[x] = y
}

func handler() *sherlock {
	if packages == nil {
		packages = make(map[string]*sherlock)
	}
	caller := caller()
	s, ok := packages[caller]
	if !ok {
		s = newSherlock()
		packages[caller] = s
	}
	return s
}

func caller() string {
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		panic(errUnrecoverable)
	}
	i := strings.LastIndex(file, "/")
	return file[:i]
}

func lookup(s *sherlock, err error, stack []byte) error {
	// search basic registry
	_, ok := s.errors[err]
	if ok {
		return err
	}
	// search map registry
	val, ok := s.mappings[err]
	if ok {
		return val
	}
	// search registered regular expressions
	str := err.Error()
	for key, _ := range s.regexps {
		ok, _ := regexp.MatchString(key, str)
		if ok {
			return err
		}
	}
	// search registered regular expression mappings
	for key, val := range s.regexpMappings {
		ok, _ := regexp.MatchString(key, str)
		if ok {
			return val
		}
	}
	// print diagnostic info to stderr and return an unexpected error
	fmt.Fprintf(os.Stderr, "Sherlock received unexpected error: %v\n", err.Error())
	fmt.Fprintf(os.Stderr, string(stack))
	return ErrUnexpected
}

func Assert(statement bool, err error) {
	if statement == false {
		panic(err)
	}
}

func Try(vals ...interface{}) {
	x := vals[len(vals)-1]
	if x != nil {
		err, ok := x.(error)
		if !ok {
			return
		}
		panic(lookup(handler(), err, debug.Stack()))
	}
}

func Check(err error) {
	if err != nil {
		panic(lookup(handler(), err, debug.Stack()))
	}
}

func Catch(err *error) {
	r := recover()
	if r != nil {
		x, ok := r.(error)
		if ok {
			*err = x
		}
	}
}
