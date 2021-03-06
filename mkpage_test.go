//
// mkpage is a thought experiment in a light weight template and markdown processor.
//
// @author R. S. Doiel, <rsdoiel@gmail.com>
//
// Copyright 2017 R. S. Doiel
//
// Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
//
package mkpage

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"text/template"

	// 3rd Party Packages
	"github.com/russross/blackfriday"
)

func TestResolveData(t *testing.T) {
	checkMap := func(ky string, expected string, m map[string]interface{}) error {
		if val, ok := m[ky]; ok == true {
			switch vv := val.(type) {
			case string:
				s := fmt.Sprintf("%s", val)
				if strings.Compare(expected, s) == 0 {
					return nil
				}
				return fmt.Errorf("expected %q, found %q, %d", expected, s, strings.Compare(expected, s))
			default:
				return fmt.Errorf("expected %s, found type %b, %s", expected, vv, val)
			}
		} else {
			return fmt.Errorf("expected %s, missing %s", expected, ky)
		}
		return nil
	}

	keyValues := map[string]string{
		"Hello":   "text:Hi there!",
		"Hi":      "markdown:*Hi there!*",
		"Nav":     path.Join("testdata", "nav.md"),
		"Content": path.Join("testdata", "content.md"),
		"Weather": "http://forecast.weather.gov/MapClick.php?lat=13.4712&lon=144.7496&FcstType=json",
	}
	data, err := ResolveData(keyValues)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := checkMap("Hello", "Hi there!", data); err != nil {
		t.Error(err)
		t.FailNow()
	}
	if err := checkMap("Hi", "<p><em>Hi there!</em></p>\n", data); err != nil {
		t.Error(err)
		t.FailNow()
	}

	src, err := ioutil.ReadFile(path.Join("testdata", "nav.md"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expected := string(blackfriday.MarkdownCommon(src))

	if err := checkMap("Nav", expected, data); err != nil {
		t.Error(err)
		t.FailNow()
	}

	src, err = ioutil.ReadFile(path.Join("testdata", "content.md"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expected = string(blackfriday.MarkdownCommon(src))

	if err := checkMap("Content", expected, data); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if _, ok := data["Weather"]; ok == false {
		t.Error("Expected a JSON blob for weather")
		t.FailNow()
	}
}

func TestMakePage(t *testing.T) {
	checkForString := func(src, target string) bool {
		if strings.Contains(src, target) == false {
			t.Errorf("expected %q in %s", target, src)
			return false
		}
		return true
	}

	src := `
Hello {{.hello}}

Nav: {{.nav}}

Content: {{.content}}

Weather: {{.weather.data.text}}
`

	keyValues := map[string]string{
		"hello":   "text:Hi there!",
		"nav":     path.Join("testdata", "nav.md"),
		"content": path.Join("testdata", "content.md"),
		"weather": "http://forecast.weather.gov/MapClick.php?lat=13.4712&lon=144.7496&FcstType=json",
	}

	tmpl := template.Must(template.New("test.tmpl").Parse(src))
	out, err := MakePageString(tmpl, keyValues)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	checkForString(out, "Hi there!")
	checkForString(out, "<ul>")
}

func TestRelativeDocPath(t *testing.T) {
	pathOK := func(t *testing.T, expected, found string) {
		if strings.Compare(expected, found) != 0 {
			t.Errorf("expected %q, found %q", expected, found)
		}
	}
	target := "css/sites.css"
	src := "index.html"
	pathOK(t, target, RelativeDocPath(src, target))

	src = "module/index.html"
	pathOK(t, "../css/sites.css", RelativeDocPath(src, target))

	src = "modules/chapter-01/"
	pathOK(t, "../../css/sites.css", RelativeDocPath(src, target))

	src = "modules/chapter-01/index.html"
	pathOK(t, "../../css/sites.css", RelativeDocPath(src, target))
}

func TestBasic(t *testing.T) {
	src := `

# Opening Slide

## This is an introduction

+ me, <me@example.org>
+ [mygitrepo.org](http://mygitrepo.org)
+ February 31, 2016, Somebody's University, Someplace

--

# Slide Two

| col A | col B | Col C |
|-------|-------|-------|
| A     | One   | 1     |

--

## Slide Three

This is slide three, just a random paragraph of text. Blah, blah, blah, blah, blah, blah, boink!


`

	tmpl, err := template.New("test.tmpl").Parse(DefaultSlideTemplateSource)
	if err != nil {
		t.Errorf("Can't parse DefaultTemplateSource templates %s", err)
		t.FailNow()
	}

	titles := []string{
		"<h1>Opening Slide</h1>",
		"<h1>Slide Two</h1>",
		"<h2>Slide Three</h2>",
	}

	slides := MarkdownToSlides("test.html", "This is just a test", "", "", []byte(src))
	if len(slides) != 3 {
		t.Errorf("Was expected three slides %+v\n", slides)
	}

	for i, slide := range slides {
		s, err := MakeSlideString(tmpl, slide)
		if err != nil {
			t.Errorf("MakeSlideString() failed %d - %s", i, err)
		}
		if strings.Contains(s, titles[i]) == false {
			t.Errorf("Expected %q in slide %d -> %s", titles[i], i, s)
		}
	}
}

func TestGrep(t *testing.T) {
	src := `
# This is some article

by Jane Doe 3001-01-01

It was some New Years day...
`
	expected := `by Jane Doe 3001-01-01`
	result := Grep(BylineExp, src)
	if expected != result {
		t.Errorf("expected %q, got %q for Grep() for byline", expected, result)
	}
	expected = `# This is some article`
	result = Grep(TitleExp, src)
	if expected != result {
		t.Errorf("expected %q, got %q for Grep() for title", expected, result)
	}
}
