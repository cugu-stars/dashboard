package main

import (
	"bytes"
	"html/template"
)

func icon(project Project) string {
	switch project.Hoster {
	case "github.com":
		return "![GitHub Icon](style/github.png)"
	case "gitlab.com":
		return "![Gitlab Icon](style/gitlab.png)"
	default:
		return ""
	}
}

func link(project Project) string {
	return "[" + project.Name + "](" + project.URL + ")"
}

func markdownBadge(badge, link string, condition func(p Project) bool) func(Project) string {
	return func(project Project) string {
		buf := &bytes.Buffer{}
		if condition == nil || condition(project) {
			buf.WriteString("[![badge](")
			badgeurl := template.Must(template.New("badge").Parse(badge))
			badgeurl.Execute(buf, project)
			buf.WriteString(")](")
			linkurl := template.Must(template.New("link").Parse(link))
			linkurl.Execute(buf, project)
			buf.WriteString(")")
		}
		return buf.String()
	}
}
