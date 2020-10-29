package badge

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/enfipy/locker"
	"github.com/narqo/go-badge"
	"github.com/xanzy/go-gitlab"
)

var isGitLab = func(p Project) bool { return p.Hoster == "gitlab.com" || p.IsGitlab }

func InitGitLabBadges(gitlabAccessToken string) {
	gitlabProject := NewGitLabProject(gitlabAccessToken)
	badges["gitlab-branches"] = gitlabProject.branches
	badges["gitlab-coverage"] = markdownBadge("{{.URL}}/badges/master/coverage.svg", "{{.URL}}/-/jobs/artifacts/master/file/coverage.html?job=unittests", isGitLab)
	badges["gitlab-forks"] = gitlabProject.forks
	badges["gitlab-issues"] = gitlabProject.issues
	badges["gitlab-lastcommit"] = gitlabProject.lastcommit
	badges["gitlab-mergerequests"] = gitlabProject.mergerequests
	badges["gitlab-pipeline"] = markdownBadge("{{.URL}}/badges/master/pipeline.svg", "{{.URL}}/pipelines", isGitLab)
	badges["gitlab-size"] = gitlabProject.size
	badges["gitlab-stars"] = gitlabProject.stars
	badges["gitlab-version"] = gitlabProject.tag
	badges["gitlab-visibility"] = gitlabProject.visibility
}

type GitLabProject struct {
	Clients           map[string]*gitlab.Client
	locker            *locker.Locker
	repositoryCache   sync.Map
	gitlabAccessToken string
}

var clientLock sync.Mutex
var Insecure = false

func (b *GitLabProject) GetClient(name string) (*gitlab.Client, error) {
	clientLock.Lock()
	defer clientLock.Unlock()

	if client, ok := b.Clients[name]; ok {
		return client, nil
	}

	transCfg := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: Insecure},
		TLSHandshakeTimeout: 10 * time.Second,
	}
	httpClient := &http.Client{Transport: transCfg, Timeout: 10 * time.Second}
	c := gitlab.NewClient(httpClient, b.gitlabAccessToken)
	err := c.SetBaseURL("https://" + name)
	if err != nil {
		return nil, err
	}

	b.Clients[name] = c

	return c, nil
}

func NewGitLabProject(gitlabAccessToken string) *GitLabProject {
	return &GitLabProject{
		gitlabAccessToken: gitlabAccessToken,
		Clients:           map[string]*gitlab.Client{},
		locker:            locker.Initialize(),
	}
}

func (b *GitLabProject) GetProject(project Project) (*gitlab.Project, error) {
	if !isGitLab(project) {
		return nil, errors.New("not a GitLab project")
	}

	client, err := b.GetClient(project.Hoster)
	if err != nil {
		return nil, err
	}

	b.locker.Lock("repo" + project.URL)
	defer b.locker.Unlock("repo" + project.URL)

	loadedProject, ok := b.repositoryCache.Load(project.URL)
	if !ok {
		t := true
		options := &gitlab.GetProjectOptions{Statistics: &t}
		id := strings.Trim(project.Namespace+"/"+project.Name, "/")
		gitlabProject, _, err := client.Projects.GetProject(id, options)
		if err != nil {
			return nil, err
		}
		b.repositoryCache.Store(project.URL, gitlabProject)
		return gitlabProject, nil
	}
	return loadedProject.(*gitlab.Project), nil
}

func (b *GitLabProject) mergerequests(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	state := "opened"
	options := &gitlab.ListProjectMergeRequestsOptions{
		State: &state,
	}

	client, _ := b.GetClient(project.Hoster)

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	_, response, err := client.MergeRequests.ListProjectMergeRequests(id, options)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "mergerequests", "merge requests", "Error", badge.ColorLightgrey, project.URL, err)
	}

	color := badge.ColorBrightgreen
	if response.TotalItems > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "mergerequests", "merge requests", fmt.Sprintf("%d", response.TotalItems), color, project.URL+"/-/merge_requests", nil)
}

func (b *GitLabProject) branches(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	options := &gitlab.ListBranchesOptions{}
	client, _ := b.GetClient(project.Hoster)

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	_, response, err := client.Branches.ListBranches(id, options)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "branches", "branches", "Error", badge.ColorLightgrey, project.URL, err)
	}

	color := badge.ColorBrightgreen
	if response.TotalItems > 1 {
		color = badge.ColorGreen
	}
	if response.TotalItems > 2 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "branches", "branches", fmt.Sprintf("%d", response.TotalItems), color, project.URL+"/-/branches", nil)
}

func (b *GitLabProject) tag(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	options := &gitlab.ListTagsOptions{}
	client, _ := b.GetClient(project.Hoster)

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	tags, _, err := client.Tags.ListTags(id, options)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "tag", "tag", "Error", badge.ColorLightgrey, project.URL, err)
	}
	if len(tags) == 0 {
		return nil
	}
	tag := tags[0]
	return svgBadge(project.Hoster, project.Name, "tag", "tag", tag.Name, badge.ColorBlue, project.URL+"/-/tags", nil)
}

func (b *GitLabProject) issues(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	gitlabProject, err := b.GetProject(project)
	if err != nil {
		return errorBadge("issues", project, err)
	}

	color := badge.ColorBrightgreen
	if gitlabProject.OpenIssuesCount > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "issues", "issues", fmt.Sprintf("%d", gitlabProject.OpenIssuesCount), color, project.URL+"/-/issues", nil)
}

func errorBadge(name string, project Project, err error) *Badge {
	return svgBadge(project.Hoster, project.Name, name, name, "Error", badge.ColorLightgrey, project.URL, err)
}

func (b *GitLabProject) lastcommit(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	gitlabProject, err := b.GetProject(project)
	if err != nil {
		return errorBadge("issues", project, err)
	}
	color := badge.ColorRed
	switch {
	case time.Now().Add(-time.Hour * 24 * 30).Before(*gitlabProject.LastActivityAt):
		color = badge.ColorBrightgreen
	case time.Now().Add(-time.Hour * 24 * 60).Before(*gitlabProject.LastActivityAt):
		color = badge.ColorGreen
	case time.Now().Add(-time.Hour * 24 * 185).Before(*gitlabProject.LastActivityAt):
		color = badge.ColorYellowgreen
	case time.Now().Add(-time.Hour * 24 * 365).Before(*gitlabProject.LastActivityAt):
		color = badge.ColorYellow
	case time.Now().Add(-time.Hour * 24 * 730).Before(*gitlabProject.LastActivityAt):
		color = badge.ColorOrange
	}
	return svgBadge(project.Hoster, project.Name, "last", "last commit", humanize.Time(*gitlabProject.LastActivityAt), color, project.URL+"/-/commits", nil)
}

func (b *GitLabProject) stars(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	gitlabProject, err := b.GetProject(project)
	if err != nil {
		return errorBadge("issues", project, err)
	}
	return svgBadge(project.Hoster, project.Name, "stars", "stars", fmt.Sprint(gitlabProject.StarCount), badge.ColorBlue, project.URL+"/-/starrers", nil)
}

func (b *GitLabProject) visibility(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	gitlabProject, err := b.GetProject(project)
	if err != nil {
		return errorBadge("issues", project, err)
	}
	color := badge.ColorBlue
	switch gitlabProject.Visibility {
	case gitlab.PrivateVisibility:
		color = badge.ColorYellow
	case gitlab.PublicVisibility:
		color = badge.ColorGreen
	}
	return svgBadge(project.Hoster, project.Name, "visibility", "visibility", string(gitlabProject.Visibility), color, project.URL, nil)
}

func (b *GitLabProject) forks(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	gitlabProject, err := b.GetProject(project)
	if err != nil {
		return errorBadge("issues", project, err)
	}
	return svgBadge(project.Hoster, project.Name, "fork", "Fork", fmt.Sprint(gitlabProject.ForksCount), badge.ColorBlue, project.URL+"/-/forks", nil)
}

func (b *GitLabProject) size(project Project) *Badge {
	if !isGitLab(project) {
		return nil
	}

	gitlabProject, err := b.GetProject(project)
	if err != nil {
		return errorBadge("issues", project, err)
	}

	color := badge.ColorBrightgreen
	switch {
	case gitlabProject.Statistics.RepositorySize > 1024*1024*100:
		color = badge.ColorRed
	case gitlabProject.Statistics.RepositorySize > 1024*1024*50:
		color = badge.ColorOrange
	case gitlabProject.Statistics.RepositorySize > 1024*1024*10:
		color = badge.ColorYellow
	case gitlabProject.Statistics.RepositorySize > 1024*1024*5:
		color = badge.ColorYellowgreen
	case gitlabProject.Statistics.RepositorySize > 1024*1024:
		color = badge.ColorGreen
	}
	return svgBadge(project.Hoster, project.Name, "reposize", "repo size", humanize.Bytes(uint64(gitlabProject.Statistics.RepositorySize)), color, project.URL, nil)
}
