package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
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
	AzureOrganization string   `yaml:"azure-organization,omitempty"`
	AzureProject      string   `yaml:"azure-project,omitempty"`
	AzureDefinitionID string   `yaml:"azure-definition-id,omitempty"`
	GoImportPath      string   `yaml:"goimportpath,omitempty"`
	Workflow          string   `yaml:"workflow,omitempty"`
	URL               string   `yaml:"url,omitempty"`
	Disable           []string `yaml:"disable,omitempty"`
	Enable            []string `yaml:"enable,omitempty"`
}

var (
	GITLAB_ACCESS_TOKEN string = ""
	GITHUB_ACCESS_TOKEN string = ""
)

func main() {
	flag.StringVar(&GITLAB_ACCESS_TOKEN, "gitlab", LookupEnvOrString("GITLAB_ACCESS_TOKEN", GITLAB_ACCESS_TOKEN), "gitlab access token")
	flag.StringVar(&GITHUB_ACCESS_TOKEN, "github", LookupEnvOrString("GITHUB_ACCESS_TOKEN", GITHUB_ACCESS_TOKEN), "github access token")

	flag.Parse()

	if GITHUB_ACCESS_TOKEN == "" || GITLAB_ACCESS_TOKEN == "" {
		log.Fatalf("Token not defined '%s' '%s'", GITHUB_ACCESS_TOKEN, GITLAB_ACCESS_TOKEN)
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config, err := parseInput()
	if err != nil {
		return err
	}
	setup()

	var wg sync.WaitGroup
	var badges sync.Map
	for _, category := range config.Categories {
		for pID, project := range category.Projects {
			project, err = parseProject(project)
			if err != nil {
				log.Println(err)
			}
			category.Projects[pID] = project

			for _, column := range config.Table {
				for _, badgeName := range column.Enabled {
					wg.Add(1)
					go func(category Category, project Project, badgeName string) {
						defer wg.Done()
						if !contains(project.Disable, badgeName) && !contains(project.Disable, column.Name) {
							if renderFunc, ok := cells[badgeName]; ok {
								badges.Store(category.Name+project.URL+badgeName, renderFunc(project))
							} else {
								log.Println(badgeName + " badge missing")
							}
						}
					}(category, project, badgeName)
				}
				for _, badgeName := range column.Disabled {
					wg.Add(1)
					go func(category Category, project Project, badgeName string) {
						defer wg.Done()
						if contains(project.Enable, badgeName) {
							if renderFunc, ok := cells[badgeName]; ok {
								badges.Store(category.Name+project.URL+badgeName, renderFunc(project))
							} else {
								log.Println(badgeName + " badge missing")
							}
						}
					}(category, project, badgeName)
				}
			}
		}
	}
	wg.Wait()

	buf := ""
	for i, category := range config.Categories {
		buf += "| **"  + category.Name + "** " + strings.Repeat("|", len(config.Table)) + "\n"
		if i == 0 {
			buf += createHeader(config.Table)
		}
		for _, project := range category.Projects {
			buf += "|"
			for _, column := range config.Table {
				for _, badgeName := range append(column.Enabled, column.Disabled...) {
					if b, ok := badges.Load(category.Name + project.URL + badgeName); ok {
						buf += b.(string)
					}
				}
				buf += "|"
			}
			buf += "\n"
		}
	}

	ioutil.WriteFile("index.md", []byte(buf), 0666)
	return createHTML("index", []byte(buf))
}

func parseInput() (config YAML, err error) {
	yamlFile, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		return
	}
	return config, yaml.Unmarshal(yamlFile, &config)
}

func createHeader(table []Column) (buf string) {
	seperationRow := ""
	for _, column := range table {
		buf += "| " + column.Name + " "
		seperationRow += "| --- "
	}
	return seperationRow + " |\n" + buf + " |\n"
}

func parseProject(project Project) (Project, error) {
	u, err := url.Parse(project.URL)
	if err != nil {
		return Project{}, err
	}
	project.Hoster = u.Host
	project.Namespace = strings.TrimLeft(path.Dir(u.Path), "/")
	project.Name = path.Base(u.Path)
	if project.GoImportPath == "" {
		project.GoImportPath = project.Hoster + "/" + project.Namespace + "/" + project.Name
	}
	return project, nil
}

func createHTML(name string, md []byte) error {
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank | html.CompletePage,
		CSS:   "style/style.css",
	})

	output := markdown.ToHTML(md, nil, renderer)
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

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
