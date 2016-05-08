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

func parseSplice(in, pathSep string) ([]splicePiece, error) {
	var pieces []splicePiece
	content := in
	off := 0
	for len(content) > 0 {
		idx := strings.Index(content[off:], "${")
		if idx < 0 {
			pieces = append(pieces, stringPiece(content))
			break
		}

		idx += off
		off = idx + 2
		// if '$${', ignore and continue parsing
		if idx > 0 && content[idx-1] == '$' {
			continue
		}

		// found start of variable, store passed content into pieces
		str := content[:idx]
		if str != "" {
			pieces = append(pieces, stringPiece(str))
		}

		// find variable end:
		end := strings.Index(content[off:], "}")
		if end < 0 {
			// err, found variable start without end
			return nil, errUnterminatedBrace
		}

		// get variable content indices + update offset
		start := off
		end += off
		off = end + 1

		// parse variable
		path := parsePath(content[start:end], pathSep)
		pieces = append(pieces, newReference(path))

		// reset string and parse offset
		content = content[off:]
		off = 0
	}

	return pieces, nil
}
