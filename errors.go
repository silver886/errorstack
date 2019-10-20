package errorstack

import (
	"fmt"
	"io"
	"strings"
)

// Stack is a stack of error.
type Stack struct {
	errs []error
}

// New returns an error stack.
func New(errs ...error) *Stack {
	return &Stack{
		errs,
	}
}

// Error is the implementation of error.
func (s *Stack) Error() string {
	var buf strings.Builder
	s.walkPrint(&buf, 'v', ": ", func(w io.Writer, v rune, e error) {
		io.WriteString(w, e.Error())
	})
	return buf.String()
}

// Format is the implementation of formater.
func (s *Stack) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			s.walkPrint(state, verb, "\n", func(w io.Writer, v rune, e error) {
				if formatter, ok := e.(fmt.Formatter); ok {
					if state, ok := w.(fmt.State); ok {
						formatter.Format(state, v)
					}
				} else {
					io.WriteString(w, e.Error())
				}
			})
			return
		}
		fallthrough
	case 's':
		s.walkPrint(state, verb, ": ", func(w io.Writer, v rune, e error) {
			io.WriteString(w, e.Error())
		})
	case 'q':
		s.walkPrint(state, verb, ": ", func(w io.Writer, v rune, e error) {
			fmt.Fprintf(w, "%q", e)
		})
	}
}
