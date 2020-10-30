package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/markbates/pkger"

	"github.com/cugu/dashboard/badge"
	"github.com/xanzy/go-gitlab"
)

//go:generate pkger

type Config struct {
	Table      []Column   `yaml:"table,omitempty"`
	Categories []Category `yaml:"categories,omitempty"`
	StaticPath string     `yaml:"staticpath,omitempty"`
}

type Column struct {
	Name     string   `yaml:"name,omitempty"`
	Enabled  []string `yaml:"enabled,omitempty"`
	Disabled []string `yaml:"disabled,omitempty"`
}

type Category struct {
	Name     string          `yaml:"name,omitempty"`
	Projects []badge.Project `yaml:"projects,omitempty"`
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	
	gitlabAccessToken := flag.String("gitlab", LookupEnvOrString("GITLAB_ACCESS_TOKEN"), "GitLab access token")
	gitlabPushBadges := flag.Bool("gitlab-push-badges", strings.ToLower(LookupEnvOrString("GITLAB_PUSH_BADGES")) == "true", "push badges to GitLab")
	githubAccessToken := flag.String("github", LookupEnvOrString("GITHUB_ACCESS_TOKEN"), "GitHub access token")
	flag.Parse()

	if *githubAccessToken == "" {
		log.Println("GitHub token not defined. GitHub Badges will not be available.")
	} else {
		badge.InitGitHubBadges(*githubAccessToken)
	}

	if *gitlabAccessToken == "" {
		log.Println("GitLab token not defined. GitLab Badges will not be available.")
	} else {
		badge.InitGitLabBadges(*gitlabAccessToken)
	}

	badge.InitDefaultBadges()
	badge.InitAzureBadges()
	badge.InitMissingFileBadges()
	badge.InitExternalCommandBadges()
	badge.Insecure = true

	if err := run(*gitlabAccessToken, *githubAccessToken, *gitlabPushBadges); err != nil {
		log.Fatal(err)
	}
}

func run(gitlabAccessToken, githubAccessToken string, gitlabPushBadges bool) error {
	config, err := parseInput()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var badges sync.Map
	for _, category := range config.Categories {
		for pID, project := range category.Projects {
			project, err = parseProject(project, gitlabAccessToken, githubAccessToken)
			if err != nil {
				log.Println(err)
			}
			category.Projects[pID] = project

			for _, column := range config.Table {
				for _, badgeName := range column.Enabled {
					wg.Add(1)
					go func(category Category, project badge.Project, badgeName string) {
						defer wg.Done()
						if !contains(project.Disable, badgeName) && !contains(project.Disable, column.Name) {
							if renderFunc, ok := badge.GetBadge(badgeName); ok {
								badges.Store(category.Name+project.URL+badgeName, renderFunc(project))
							} else {
								log.Println(badgeName + " badge missing")
							}
						}
					}(category, project, badgeName)
				}
				for _, badgeName := range column.Disabled {
					wg.Add(1)
					go func(category Category, project badge.Project, badgeName string) {
						defer wg.Done()
						if contains(project.Enable, badgeName) {
							if renderFunc, ok := badge.GetBadge(badgeName); ok {
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

	if waitTimeout(&wg, 10*time.Minute) {
		fmt.Println("Timed out waiting for wait group")
	} else {
		fmt.Println("Wait group finished")
	}

	err = os.MkdirAll("style", os.ModePerm)
	if err != nil {
		return err
	}
	err = pkger.Walk("/static", func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		src, err := pkger.Open(p)
		if err != nil {
			return err
		}
		defer src.Close()

		dest, err := os.Create(path.Join("style", info.Name()))
		if err != nil {
			return err
		}
		defer dest.Close()
		_, err = io.Copy(dest, src)
		return err
	})
	if err != nil {
		return err
	}

	md, err := createMarkdown(config.Categories, config.Table, &badges)
	if err != nil {
		return err
	}
	err = createHTML("index", []byte(md))
	if err != nil {
		return err
	}

	if gitlabPushBadges {
		return createGitLabBadges(config.Categories, config.Table, &badges, gitlabAccessToken, config.StaticPath)
	}
	return nil
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func createGitLabBadges(categories []Category, table []Column, badges *sync.Map, gitlabAccessToken, staticPath string) error {
	gl := badge.NewGitLabProject(gitlabAccessToken)

	var isGitLab = func(p badge.Project) bool { return p.Hoster == "gitlab.com" || p.IsGitlab }

	for _, category := range categories {
		for _, project := range category.Projects {
			if !isGitLab(project) {
				continue
			}

			id := strings.Trim(project.Namespace+"/"+project.Name, "/")

			client, _ := gl.GetClient(project.Hoster)

			// delete all old badges
			oldBadges, _, err := client.ProjectBadges.ListProjectBadges(id, &gitlab.ListProjectBadgesOptions{})
			if err != nil {
				return err
			}

			authError := false
			for _, b := range oldBadges {
				_, err = client.ProjectBadges.DeleteProjectBadge(id, b.ID)
				if err != nil {
					authError = true
					break
				}
			}
			if authError {
				continue
			}

			for _, column := range table {
				for _, badgeName := range append(column.Enabled, column.Disabled...) {
					if b, ok := badges.Load(category.Name + project.URL + badgeName); ok {

						pBadge := b.(*badge.Badge)
						if pBadge == nil {
							continue
						}

						u := pBadge.URL
						if !strings.HasPrefix(u, "http") {
							u = fmt.Sprintf(staticPath, u)
						}
						l := pBadge.Link
						if !strings.HasPrefix(l, "http") {
							l = fmt.Sprintf(staticPath, l)
						}

						opt := &gitlab.AddProjectBadgeOptions{LinkURL: &l, ImageURL: &u}
						_, _, err := client.ProjectBadges.AddProjectBadge(id, opt)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func parseProject(project badge.Project, gitlabAccessToken, githubAccessToken string) (badge.Project, error) {
	u, err := url.Parse(project.URL)
	if err != nil {
		return badge.Project{}, err
	}
	project.Hoster = u.Host

	if project.Hoster == "gitlab.com" || project.IsGitlab {
		project.Token = gitlabAccessToken
	}
	if project.Hoster == "github.com" {
		project.Token = githubAccessToken
	}

	project.Namespace = strings.TrimLeft(path.Dir(u.Path), "/")
	project.Name = path.Base(u.Path)
	if project.GoImportPath == "" {
		project.GoImportPath = project.Hoster + "/" + project.Namespace + "/" + project.Name
	}
	return project, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.EqualFold(a, e) {
			return true
		}
	}
	return false
}

func LookupEnvOrString(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return ""
}
