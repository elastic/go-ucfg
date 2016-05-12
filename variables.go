package ucfg

import (
	"bytes"
	"errors"
	"strings"
)

type reference struct {
	path cfgPath
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
)

const (
	varOpen  = "${"
	varClose = "$}"
)

func newReference(p cfgPath) *reference {
	return &reference{p}
}

func (r *reference) resolve(cfg *Config) (value, error) {
	root := cfgRoot(cfg)
	if root == nil {
		return nil, ErrMissing
	}
	return r.path.GetValue(root)
}

func (r *reference) piece(cfg *Config) (string, error) {
	v, err := r.resolve(cfg)
	if err != nil {
		return "", err
	}
	return v.toString()
}

func (s stringPiece) piece(cfg *Config) (string, error) {
	return string(s), nil
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

func lexer(in string) (<-chan string, <-chan error) {
	lex := make(chan string, 1)
	errors := make(chan error, 1)

	go func() {
		off := 0
		content := in

		defer func() {
			if len(content) > 0 {
				lex <- content
			}
			close(lex)
			close(errors)
		}()

		for len(content) > 0 {
			idx := strings.Index(content[off:], "${")
			if idx < 0 {
				return
			}

			idx += off
			off = idx + 2
			if idx > 0 && content[idx-1] == '$' {
				// if '$${', ignore and continue parsing
				continue
			}

			// found start of variable, store passed content into pieces
			if str := content[:idx]; str != "" {
				lex <- str
			}

			// find variable end:
			end := strings.Index(content[off:], "}")
			if end < 0 {
				// err, found variable start without end
				errors <- errUnterminatedBrace
				return
			}

			// get variable content indices + update offset
			start := off
			end += off
			off = end + 1

			// pass variable
			lex <- varOpen
			lex <- content[start:end]
			lex <- varClose

			content = content[off:]
			off = 0
		}
	}()

	return lex, errors
}

func parseSplice(in, pathSep string) ([]splicePiece, error) {
	lex, errors := lexer(in)

	// lexer co-routine
	var pieces []splicePiece
	isvar := false
	for sym := range lex {
		// process symbol
		switch sym {
		case varOpen:
			isvar = true
		case varClose:
			isvar = false
		default:
			if isvar {
				path := parsePath(sym, pathSep)
				pieces = append(pieces, newReference(path))
			} else {
				pieces = append(pieces, stringPiece(sym))
			}
		}
	}

	err := <-errors
	if err != nil {
		return nil, err
	}
	return pieces, nil
}
