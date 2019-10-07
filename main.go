package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"strings"

	"gopkg.in/russross/blackfriday.v2"
	"gopkg.in/yaml.v2"
)

type YAML struct {
	Table      []Column   `yaml:"table,omitempty"`
	Categories []Category `yaml:"categories,omitempty"`
}

type Column struct {
	Name     string   `yaml:"name,omitempty"`
	Enabled  []string `yaml:"enabled,omitempty"`
	Disabled []string `yaml:"disabled,omitempty"`
}

type Category struct {
	Name     string    `yaml:"name,omitempty"`
	Projects []Project `yaml:"projects,omitempty"`
}

type Project struct {
	Hoster            string   `yaml:"-"`
	Namespace         string   `yaml:"-"`
	Name              string   `yaml:"-"`
	AzureDefinitionID string   `yaml:"azure-definition-id,omitempty"`
	GoImportPath      string   `yaml:"goimportpath,omitempty"`
	URL               string   `yaml:"url,omitempty"`
	Disable           []string `yaml:"disable,omitempty"`
	Enable            []string `yaml:"enable,omitempty"`
}

var (
	GITLAB_ACCESS_TOKEN string = ""
)

func main() {
	flag.StringVar(&GITLAB_ACCESS_TOKEN, "gitlab", LookupEnvOrString("GITLAB_ACCESS_TOKEN", GITLAB_ACCESS_TOKEN), "gitlab access token")

	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	config, err := parseInput()
	if err != nil {
		return err
	}

	for _, category := range config.Categories {
		buf := createNav(config.Categories) + "## " + category.Name + "\n\n" + createHeader(config.Table)
		for _, project := range category.Projects {
			project, err = parseProject(project)
			if err != nil {
				return err
			}

			for _, column := range config.Table {
				buf += createCell(project, column)
			}
			buf += "|\n"
		}
		ioutil.WriteFile(category.Name+".md", []byte(buf), 0666)
		if err := createHTML(category.Name, []byte(buf)); err != nil {
			return err
		}
	}
	return nil
}

func parseInput() (config YAML, err error) {
	yamlFile, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		return
	}
	return config, yaml.Unmarshal(yamlFile, &config)
}

func createNav(categories []Category) (buf string) {
	links := []string{}
	for _, categorylink := range categories {
		links = append(links, "["+categorylink.Name+"]("+categorylink.Name+".html)")
	}
	return strings.Join(links, " - ") + "\n\n"
}

func createHeader(table []Column) (buf string) {
	seperationRow := ""
	for _, column := range table {
		buf += "| " + column.Name + " "
		seperationRow += "| --- "
	}
	return buf + " |\n" + seperationRow + " |\n"
}

func parseProject(project Project) (Project, error) {
	u, err := url.Parse(project.URL)
	if err != nil {
		return Project{}, err
	}
	project.Hoster = u.Host
	project.Namespace = path.Dir(u.Path)
	project.Name = path.Base(u.Path)
	if project.GoImportPath == "" {
		project.GoImportPath = project.Hoster + project.Namespace + "/" + project.Name
	}
	return project, nil
}

func createCell(project Project, column Column) (buf string) {
	for _, badgeName := range column.Enabled {
		if !contains(project.Disable, badgeName) && !contains(project.Disable, column.Name) {
			buf += cells[badgeName].Render(column.Name, project)
		}
	}
	for _, badgeName := range column.Disabled {
		if contains(project.Enable, badgeName) {
			buf += cells[badgeName].Render(column.Name, project)
		}
	}
	return "|" + buf
}

func createHTML(name string, markdown []byte) error {
	params := blackfriday.HTMLRendererParameters{
		Flags: blackfriday.CommonHTMLFlags | blackfriday.CompletePage,
		CSS:   "style/style.css",
	}
	renderer := blackfriday.NewHTMLRenderer(params)
	output := blackfriday.Run(markdown, blackfriday.WithExtensions(blackfriday.CommonExtensions), blackfriday.WithRenderer(renderer))
	return ioutil.WriteFile(name+".html", output, 0666)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}
