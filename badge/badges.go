package badge

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/narqo/go-badge"
)

type Badge struct {
	URL   string `yaml:"url,omitempty"`
	Link  string `yaml:"link,omitempty"`
	Title string `yaml:"title,omitempty"`
	Error error  `yaml:"error,omitempty"`
}

func (b *Badge) ToMarkdown() string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("[![badge](%s)](%s)<!--%s-->", b.URL, b.Link, b.Error)
}

type Project struct {
	Hoster            string   `yaml:"-"`
	Namespace         string   `yaml:"-"`
	Name              string   `yaml:"-"`
	AzureOrganization string   `yaml:"azure-organization,omitempty"`
	AzureProject      string   `yaml:"azure-project,omitempty"`
	AzureDefinitionID string   `yaml:"azure-definition-id,omitempty"`
	GoImportPath      string   `yaml:"goimportpath,omitempty"`
	Workflow          string   `yaml:"workflow,omitempty"`
	URL               string   `yaml:"url,omitempty"`
	Disable           []string `yaml:"disable,omitempty"`
	Enable            []string `yaml:"enable,omitempty"`
	Token             string   `yaml:"token,omitempty"`
	IsGitlab          bool     `yaml:"gitlab,omitempty"`
}

type badgeCreation func(Project) *Badge

var badges = map[string]badgeCreation{}

func GetBadge(name string) (badgeCreation, bool) {
	val, ok := badges[name]
	return val, ok
}

func svgBadge(hoster, projectname, name, left, right string, color badge.Color, url string, e error) *Badge {
	if len(right) > 40 {
		right = right[:35]
	}
	b, err := badge.RenderBytes(left, right, color)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(filepath.Join("badges", hoster, projectname), 0777)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filepath.Join("badges", hoster, projectname, name+".svg"), bytes.ReplaceAll(b, []byte("\n"), []byte("")), 0666)
	if err != nil {
		panic(err)
	}

	return &Badge{
		URL:   fmt.Sprintf("badges/%s/%s/%s.svg", hoster, projectname, name),
		Link:  url,
		Title: name,
		Error: e,
	}
}

func markdownBadge(badge, link string, condition func(p Project) bool) badgeCreation {
	return func(project Project) *Badge {
		if condition == nil || condition(project) {
			o := Badge{}

			badgebuf, linkbuf := &bytes.Buffer{}, &bytes.Buffer{}
			badgeurl := template.Must(template.New("badge").Parse(badge))
			err := badgeurl.Execute(badgebuf, project)
			if err != nil {
				log.Fatal(err)
			}
			o.URL = badgebuf.String()

			linkurl := template.Must(template.New("link").Parse(link))
			err = linkurl.Execute(linkbuf, project)
			if err != nil {
				log.Fatal(err)
			}
			o.Link = linkbuf.String()
			return &o
		}
		return nil
	}
}
