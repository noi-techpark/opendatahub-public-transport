// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package netex

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"
)

const skipDoc = `<?xml version="1.0"?>
<root>
  <keep a="1">text</keep>
  <drop attr="x > y" other='a/>b'>
    <!-- </drop> <drop> -->
    <![CDATA[ </drop> <x> ]]>
    <?pi </drop> ?>
    <drop><drop/>nested same name</drop>
    <empty/>
    <empty attr="/" />
  </drop>
  <keep>after</keep>
  <drop/>
  <drop attr="v"/>
  <ns:drop xmlns:ns="urn:x"><ns:inner>t</ns:inner></ns:drop>
  <keep>end</keep>
</root>`

// trace records the tokens a reader sees when every top-level <drop> child of
// <root> is skipped with skip.
func trace(t *testing.T, next func() (xml.Token, error), skip func() error) []string {
	t.Helper()
	var out []string
	depth := 0
	for {
		tok, err := next()
		if err != nil {
			return append(out, "err:"+err.Error())
		}
		switch tt := tok.(type) {
		case xml.StartElement:
			depth++
			out = append(out, "<"+tt.Name.Local)
			if depth == 2 && tt.Name.Local == "drop" {
				if err := skip(); err != nil {
					t.Fatalf("skip: %v", err)
				}
				depth--
				out = append(out, "skipped")
			}
		case xml.EndElement:
			depth--
			out = append(out, "</"+tt.Name.Local)
		case xml.CharData:
			if s := strings.TrimSpace(string(tt)); s != "" {
				out = append(out, s)
			}
		}
	}
}

func TestDecoderSkipMatchesXMLDecoder(t *testing.T) {
	ref := xml.NewDecoder(strings.NewReader(skipDoc))
	want := trace(t, ref.Token, ref.Skip)

	// Buffer sizes small enough to split every tag, comment and CDATA section.
	for _, size := range []int{1, 2, 3, 5, 7, 16, 64 << 10} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			d := NewDecoder(strings.NewReader(skipDoc))
			d.src.buf = make([]byte, size)
			got := trace(t, d.Token, d.Skip)
			if strings.Join(got, "|") != strings.Join(want, "|") {
				t.Fatalf("token trace differs\n got: %v\nwant: %v", got, want)
			}
		})
	}
}

func TestDecoderSkipTruncated(t *testing.T) {
	d := NewDecoder(strings.NewReader(`<root><drop><a>unterminated`))
	for {
		tok, err := d.Token()
		if err != nil {
			t.Fatalf("token: %v", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "drop" {
			break
		}
	}
	if err := d.Skip(); err == nil {
		t.Fatal("expected error skipping a truncated element")
	}
}
