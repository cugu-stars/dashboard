package main

import (
	"io/ioutil"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

func createHTML(name string, md []byte) error {
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank | html.CompletePage,
		CSS:   "style/style.css",
	})

	output := markdown.ToHTML(md, nil, renderer)
	return ioutil.WriteFile(name+".html", output, 0666)
}
