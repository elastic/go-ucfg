package ucfg

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type reference struct {
	Path cfgPath
}

type expansion struct {
	left, right splicePiece
	pathSep     string
}

type splice struct {
	pieces []splicePiece
}

type splicePiece interface {
	piece(cfg *Config) (string, error)
}

type stringPiece string

var (
	errUnterminatedBrace = errors.New("unterminated brace")
	errInvalidType       = errors.New("invalid type")
	errEmptyPath         = errors.New("empty path after expansion")
)

type token struct {
	typ tokenType
	val string
}

type tokenType uint16

const (
	tokOpen tokenType = iota
	tokClose
	tokSep
	tokString
)

var (
	openToken  = token{tokOpen, "${"}
	closeToken = token{tokClose, "}"}
	sepToken   = token{tokSep, ":"}
)

func newReference(p cfgPath) *reference {
	return &reference{p}
}

func (r *reference) String() string {
	return fmt.Sprintf("${%v}", r.Path)
}

func (r *reference) resolve(cfg *Config) (value, error) {
	root := cfgRoot(cfg)
	if root == nil {
		return nil, ErrMissing
	}
	return r.Path.GetValue(root)
}

func (r *reference) piece(cfg *Config) (string, error) {
	v, err := r.resolve(cfg)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", fmt.Errorf("can not resolve reference: %v", r.Path)
	}
	return v.toString()
}

func (s stringPiece) piece(cfg *Config) (string, error) {
	return string(s), nil
}

func (s *splice) String() string {
	return fmt.Sprintf("%v", s.pieces)
}

func (s *splice) piece(cfg *Config) (string, error) {
	return s.eval(cfg)
}

func (s *splice) eval(cfg *Config) (string, error) {
	buf := bytes.NewBuffer(nil)
	for _, p := range s.pieces {
		s, err := p.piece(cfg)
		if err != nil {
			return "", err
		}
		buf.WriteString(s)
	}
	return buf.String(), nil
}

func (e *expansion) String() string {
	if e.right != nil {
		return fmt.Sprintf("${%v:%v}", e.left, e.right)
	}
	return fmt.Sprintf("${%v}", e.left)
}

func (e *expansion) piece(cfg *Config) (string, error) {
	path, err := e.left.piece(cfg)
	if err == nil && path == "" {
		err = errEmptyPath
	}

	s := ""
	if err == nil {
		ref := newReference(parsePath(path, e.pathSep))
		s, err = ref.piece(cfg)
	}

	if err == nil || e.right == nil {
		return s, err
	}
	return e.right.piece(cfg)
}

func cfgRoot(cfg *Config) *Config {
	if cfg == nil {
		return nil
	}

	for {
		p := cfg.Parent()
		if p == nil {
			return cfg
		}

		cfg = p
	}
}

func parseSplice(in, pathSep string) ([]splicePiece, error) {
	lex, errs := lexer(in)
	defer func() {
		// on parser error drain lexer so go-routine won't leak
		for range lex {
		}
	}()

	pieces, perr := parseVarExp(lex, pathSep)

	// check for lexer errors
	err := <-errs
	if err != nil {
		return nil, err
	}

	// return parser result
	return pieces, perr
}

func lexer(in string) (<-chan token, <-chan error) {
	lex := make(chan token, 1)
	errors := make(chan error, 1)

	go func() {
		off := 0
		content := in

		defer func() {
			if len(content) > 0 {
				lex <- token{tokString, content}
			}
			close(lex)
			close(errors)
		}()

		strToken := func(s string) {
			if s != "" {
				lex <- token{tokString, s}
			}
		}

		varcount := 0
		for len(content) > 0 {
			idx := -1
			if varcount == 0 {
				idx = strings.IndexAny(content[off:], "$")
			} else {
				idx = strings.IndexAny(content[off:], "$:}")
			}
			if idx < 0 {
				return
			}

			idx += off
			off = idx + 1
			switch content[idx] {
			case ':':
				strToken(content[:idx])
				lex <- sepToken

			case '}':
				strToken(content[:idx])
				lex <- closeToken
				varcount--

			case '$':
				if len(content) <= off { // found '$' at end of string
					return
				}

				switch content[off] {
				case '$': // escape '$' symbol
					content = content[:off] + content[off+1:]
					continue
				case '{': // start variable
					strToken(content[:idx])
					lex <- openToken
					off++
					varcount++
				}
			}

			content = content[off:]
			off = 0
		}
	}()

	return lex, errors
}

func parseVarExp(lex <-chan token, pathSep string) ([]splicePiece, error) {
	type state struct {
		st     int
		isvar  bool
		pieces [2][]splicePiece
	}

	stLeft := 0
	stRight := 1

	stack := []state{
		state{st: stLeft},
	}

	// convert finalized parse state to splicePiece
	st2piece := func(st state) (splicePiece, error) {
		if !st.isvar {
			return nil, errors.New("fatal: processing non-variable state")
		}
		if len(st.pieces[stLeft]) == 0 {
			return nil, errors.New("empty expansion")
		}

		if st.st == stLeft && len(st.pieces[stLeft]) == 1 {
			if str, ok := st.pieces[stLeft][0].(stringPiece); ok {
				// found string piece -> parse into reference
				return newReference(parsePath(string(str), pathSep)), nil
			}
		}

		left := &splice{st.pieces[stLeft]}
		var right splicePiece
		if st.st == stRight {
			if len(st.pieces[stRight]) == 0 {
				right = stringPiece("")
			} else {
				right = &splice{st.pieces[stRight]}
			}
		}

		return &expansion{left, right, pathSep}, nil
	}

	// parser loop
	for tok := range lex {
		switch tok.typ {
		case tokOpen:
			stack = append(stack, state{st: stLeft, isvar: true})
		case tokClose:
			// pop and finalize state
			piece, err := st2piece(stack[len(stack)-1])
			stack = stack[:len(stack)-1]
			if err != nil {
				return nil, err
			}

			// append result to most recent state
			st := &stack[len(stack)-1]
			st.pieces[st.st] = append(st.pieces[st.st], piece)

		case tokSep: // switch from left to right
			st := &stack[len(stack)-1]
			if !st.isvar {
				return nil, errors.New("default separator not within expansion")
			}
			if st.st == stRight {
				return nil, errors.New("unexpected ':'")
			}
			st.st = stRight

		case tokString:
			// append raw string
			st := &stack[len(stack)-1]
			st.pieces[st.st] = append(st.pieces[st.st], stringPiece(tok.val))
		}
	}

	// validate and return final state
	if len(stack) > 1 {
		return nil, errors.New("missing '}'")
	}
	if len(stack) == 0 {
		return nil, errors.New("fatal: expansion parse state empty")
	}
	return stack[0].pieces[stLeft], nil
}
