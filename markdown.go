package main

import (
	"io/ioutil"
	"strings"
	"sync"

	"github.com/cugu/dashboard/badge"
)

func createMarkdown(categories []Category, table []Column, badges *sync.Map) (string, error) {
	buf := ""
	for i, category := range categories {
		if i == 0 {
			buf += createHeader(table)
		}
		buf += "| **" + category.Name + "** " + strings.Repeat("|", len(table)+1) + "\n"
		for _, project := range category.Projects {
			buf += "| [" + project.Name + "](" + project.URL + ") |"
			for _, column := range table {
				for _, badgeName := range append(column.Enabled, column.Disabled...) {
					if b, ok := badges.Load(category.Name + project.URL + badgeName); ok {
						buf += b.(*badge.Badge).ToMarkdown()
					}
				}
				buf += "|"
			}
			buf += "\n"
		}
	}

	return buf, ioutil.WriteFile("index.md", []byte(buf), 0666)
}

func createHeader(table []Column) string {
	buf := "| Name "
	seperationRow := "| --- "
	for _, column := range table {
		buf += "| " + column.Name + " "
		seperationRow += "| --- "
	}
	return buf + " |\n" + seperationRow + " |\n"
}
