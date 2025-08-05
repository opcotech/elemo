package main

import (
	"bytes"
	"log"
	"os"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/vanng822/go-premailer/premailer"
)

func main() {
	prem, err := premailer.NewPremailerFromFile(os.Args[1], premailer.NewOptions())
	if err != nil {
		log.Fatal(err)
	}

	input, err := prem.Transform()
	if err != nil {
		log.Fatal(err)
	}

	cssMinifier := &css.Minifier{
		Inline: true,
	}

	htmlMinifier := &html.Minifier{
		KeepConditionalComments: true,
	}

	var out bytes.Buffer

	m := minify.New()
	m.Add("text/css", cssMinifier)
	m.Add("text/html", htmlMinifier)

	if err := m.Minify("text/html", &out, bytes.NewReader([]byte(input))); err != nil {
		log.Fatal(err)
	}

	log.Printf("Write minified HTML to %s\n", os.Args[1])
	if err := os.WriteFile(os.Args[1], out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
