//
// mkpage is a thought experiment in a light weight template and markdown processor
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
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	// My package
	"github.com/rsdoiel/cli"
	"github.com/rsdoiel/mkpage"
	"github.com/rsdoiel/tmplfn"
)

var (
	usage = `USAGE: %s [OPTION] [KEY/VALUE DATA PAIRS] [TEMPLATE_FILENAMES]`

	description = `
SYNOPSIS

Using the key value pairs populate the template(s) and render to stdout.
`

	examples = `
EXAMPLE

Template

    Date: {{- .now}}

    Hello {{.name -}},
    
    The current weather is

    {{.weather}}

    Thank you

	{{.signature}}

Render the template above (i.e. myformletter.template) would be accomplished from the following
data sources--

+ "now" and "name" are strings
+ "weatherForcast" comes from a URL
+ "license" comes from a file in our local disc

That would be expressed on the command line as follows

	%s "now=text:$(date)" "name=text:Little Frieda" \
		"weather=http://forecast.weather.gov/MapClick.php?lat=13.47190933300044&lon=144.74977715100056&FcstType=json" \
		signature=testdata/signature.txt \
		testdata/myformletter.template

Golang's text/template docs can be found at 

      https://golang.org/pkg/text/template/
`

	// Standard Options
	showHelp    bool
	showVersion bool
	showLicense bool

	// Application Options
	showTemplate bool
)

func init() {
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.BoolVar(&showHelp, "help", false, "show help")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.BoolVar(&showLicense, "l", false, "show license")
	flag.BoolVar(&showLicense, "license", false, "show license")
	flag.BoolVar(&showTemplate, "t", false, "show the default template source")
	flag.BoolVar(&showTemplate, "template", false, "show the default template source")
}

func main() {
	appName := path.Base(os.Args[0])
	flag.Parse()
	args := flag.Args()

	// Configuration and command line interation
	cfg := cli.New(appName, "MKPAGE", fmt.Sprintf(mkpage.LicenseText, appName, mkpage.Version), mkpage.Version)
	cfg.UsageText = fmt.Sprintf(usage, appName)
	cfg.DescriptionText = description
	cfg.OptionsText = "OPTIONS\n"
	cfg.ExampleText = fmt.Sprintf(examples, appName)

	// Standard Options
	if showHelp == true {
		fmt.Println(cfg.Usage())
		os.Exit(0)
	}
	if showVersion == true {
		fmt.Println(cfg.Version())
		os.Exit(0)
	}
	if showLicense == true {
		fmt.Println(cfg.License())
		os.Exit(0)
	}

	// Application Options
	if showTemplate == true {
		fmt.Fprintf(os.Stdout, "%s\n", mkpage.DefaultTemplateSource)
		os.Exit(0)
	}

	var templateSources []string
	data := make(map[string]string)
	for i, arg := range args {
		if strings.Contains(arg, "=") == true {
			// Update data map
			pair := strings.SplitN(arg, "=", 2)
			if len(pair) != 2 {
				fmt.Fprintf(os.Stderr, "Can't read pair (%d) %s\n", i+1, arg)
				os.Exit(1)
			}
			data[pair[0]] = pair[1]
		} else {
			// Must be the template source
			templateSources = append(templateSources, arg)
		}
	}

	// NOTE: Now we're ready to parse and populate our template
	var (
		tmpl *template.Template
		err  error
	)
	if len(templateSources) == 0 {
		tmpl, err = template.New("default.tmpl").Funcs(tmplfn.Join(tmplfn.TimeMap, tmplfn.PageMap)).Parse(mkpage.DefaultTemplateSource)
	} else {
		tmpl, err = template.New(path.Base(templateSources[0])).Funcs(tmplfn.Join(tmplfn.TimeMap, tmplfn.PageMap)).ParseFiles(templateSources...)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Template parsing failed, %s\n", err)
		os.Exit(1)
	}
	if err := mkpage.MakePage(os.Stdout, tmpl, data); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
