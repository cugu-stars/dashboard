package badge

func InitDefaultBadges() {
	badges["icon"] = icon
	badges["travis"] = markdownBadge("https://travis-ci.org/{{.Namespace}}/{{.Name}}.svg?branch=master", "https://travis-ci.org/{{.Namespace}}/{{.Name}}", nil)
	badges["gocover"] = markdownBadge("http://gocover.io/_badge/{{.Hoster}}/{{.Namespace}}/{{.Name}}", "https://gocover.io/{{.Hoster}}/{{.Namespace}}/{{.Name}}", nil)
	badges["codecov"] = markdownBadge("https://codecov.io/gh/{{.Namespace}}/{{.Name}}/branch/master/graph/badge.svg", "https://codecov.io/gh/{{.Namespace}}/{{.Name}}", nil)
	badges["goreportcard"] = markdownBadge("https://goreportcard.com/badge/{{.GoImportPath}}", "https://goreportcard.com/report/{{.GoImportPath}}", nil)
	badges["godoc"] = markdownBadge("https://godoc.org/{{.GoImportPath}}?status.svg", "https://godoc.org/{{.GoImportPath}}", nil)
}

func icon(project Project) *Badge {
	switch project.Hoster {
	case "github.com":
		return &Badge{URL: "style/github.png", Link: project.URL, Title: "icon"}
	case "gitlab.com":
		return &Badge{URL: "style/gitlab.png", Link: project.URL, Title: "icon"}
	default:
		return nil
	}
}
