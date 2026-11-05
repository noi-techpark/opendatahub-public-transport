// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// Decoder is an xml.Decoder whose Skip jumps over a subtree on the raw bytes
// instead of tokenizing it. On large NeTEx files most bytes sit in subtrees a
// caller does not need (passing times, points in sequence), and tokenizing
// them dominates parse time.
type Decoder struct {
	*xml.Decoder
	src     *rawReader
	accepts func(entityType string) bool
}

// NewDecoder returns a Decoder reading from r that accepts every entity type.
func NewDecoder(r io.Reader) *Decoder {
	src := &rawReader{src: r, buf: make([]byte, 64<<10)}
	// xml.Decoder reads straight from an io.ByteReader without buffering ahead,
	// so src always sits exactly where the decoder stopped.
	return &Decoder{Decoder: xml.NewDecoder(src), src: src}
}

// Wants reports whether entities of entityType were requested. Parsers use it
// to skip subtrees that would only produce unrequested entities.
func (d *Decoder) Wants(entityType string) bool {
	return d.accepts == nil || d.accepts(entityType)
}

// Skip consumes the element whose StartElement Token just returned, up to and
// including its EndElement.
func (d *Decoder) Skip() error {
	if !d.src.afterEmptyTag() {
		if err := d.src.skipContent(); err != nil {
			return err
		}
	}
	tok, err := d.Token()
	if err != nil {
		return err
	}
	if _, ok := tok.(xml.EndElement); !ok {
		return fmt.Errorf("skip: expected end element, got %T", tok)
	}
	return nil
}

type rawReader struct {
	src  io.Reader
	buf  []byte
	r, w int
	err  error
}

func (s *rawReader) ReadByte() (byte, error) {
	if s.r == s.w && !s.fill() {
		return 0, s.err
	}
	b := s.buf[s.r]
	s.r++
	return b, nil
}

func (s *rawReader) Read(p []byte) (int, error) {
	if s.r == s.w && !s.fill() {
		return 0, s.err
	}
	n := copy(p, s.buf[s.r:s.w])
	s.r += n
	return n, nil
}

// fill reads more input. Unconsumed bytes and the two last consumed ones are
// kept at the front, so offsets relative to r survive and afterEmptyTag can
// look back.
func (s *rawReader) fill() bool {
	keep := max(s.r-2, 0)
	s.w = copy(s.buf, s.buf[keep:s.w])
	s.r -= keep
	if s.w == len(s.buf) {
		s.buf = append(s.buf, make([]byte, len(s.buf))...)
	}
	for s.err == nil {
		n, err := s.src.Read(s.buf[s.w:])
		s.w += n
		if err != nil {
			s.err = err
		}
		if n > 0 {
			return true
		}
	}
	return false
}

func (s *rawReader) ensure(n int) bool {
	for s.w-s.r < n {
		if !s.fill() {
			return false
		}
	}
	return true
}

func (s *rawReader) truncated() error {
	if s.err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return s.err
}

// afterEmptyTag reports whether the last consumed tag was self-closing, in
// which case xml.Decoder synthesizes the EndElement without reading input.
func (s *rawReader) afterEmptyTag() bool {
	return s.r >= 2 && s.buf[s.r-2] == '/' && s.buf[s.r-1] == '>'
}

// skipContent advances to the '<' of the end tag closing the current element.
// Well-formed XML has no raw '<' outside markup, so only tags, comments, CDATA
// and processing instructions need recognizing.
func (s *rawReader) skipContent() error {
	depth := 0
	for {
		i := bytes.IndexByte(s.buf[s.r:s.w], '<')
		if i < 0 {
			s.r = s.w
			if !s.fill() {
				return s.truncated()
			}
			continue
		}
		s.r += i
		if !s.ensure(2) {
			return s.truncated()
		}
		var err error
		switch s.buf[s.r+1] {
		case '/':
			if depth == 0 {
				return nil
			}
			depth--
			err = s.skipPast([]byte(">"))
		case '?':
			err = s.skipPast([]byte("?>"))
		case '!':
			switch {
			case s.hasPrefix("<!--"):
				err = s.skipPast([]byte("-->"))
			case s.hasPrefix("<![CDATA["):
				err = s.skipPast([]byte("]]>"))
			default:
				err = s.skipPast([]byte(">"))
			}
		default:
			var empty bool
			empty, err = s.skipStartTag()
			if !empty {
				depth++
			}
		}
		if err != nil {
			return err
		}
	}
}

func (s *rawReader) hasPrefix(p string) bool {
	return s.ensure(len(p)) && string(s.buf[s.r:s.r+len(p)]) == p
}

// skipPast consumes through the first occurrence of pat after the current '<'.
func (s *rawReader) skipPast(pat []byte) error {
	from := 1
	for {
		if i := bytes.Index(s.buf[s.r+from:s.w], pat); i >= 0 {
			s.r += from + i + len(pat)
			return nil
		}
		from = max(s.w-s.r-len(pat)+1, 1)
		if !s.fill() {
			return s.truncated()
		}
	}
}

// skipStartTag consumes a start tag and reports whether it was self-closing.
// Attribute values may contain '>', so quotes are tracked.
func (s *rawReader) skipStartTag() (bool, error) {
	var quote byte
	for j := 1; ; j++ {
		if s.r+j >= s.w && !s.ensure(j+1) {
			return false, s.truncated()
		}
		switch b := s.buf[s.r+j]; {
		case quote != 0:
			if b == quote {
				quote = 0
			}
		case b == '"' || b == '\'':
			quote = b
		case b == '>':
			empty := s.buf[s.r+j-1] == '/'
			s.r += j + 1
			return empty, nil
		}
	}
}
