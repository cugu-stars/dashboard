package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/enfipy/locker"
	"github.com/narqo/go-badge"
	"github.com/xanzy/go-gitlab"
)

type GitLabProject struct {
	client          *gitlab.Client
	locker          *locker.Locker
	repositoryCache sync.Map
}

func NewGitLabProject() *GitLabProject {
	return &GitLabProject{
		client: gitlab.NewClient(nil, GITLAB_ACCESS_TOKEN),
		locker: locker.Initialize(),
	}
}

func (b *GitLabProject) getProject(project Project) (*gitlab.Project, *string) {
	if !isGitLab(project) {
		s := ""
		return nil, &s
	}

	b.locker.Lock("repo" + project.URL)
	defer b.locker.Unlock("repo" + project.URL)

	loadedProject, ok := b.repositoryCache.Load(project.URL)
	if !ok {
		t := true
		options := &gitlab.GetProjectOptions{Statistics: &t}
		id := strings.Trim(project.Namespace+"/"+project.Name, "/")
		gitlabProject, _, err := b.client.Projects.GetProject(id, options)
		if err != nil {
			badge := svgBadge(project.Hoster, project.Name, "gitlab", "gitlab", "Error", badge.ColorLightgrey, project.URL, err)
			return nil, &badge
		}
		b.repositoryCache.Store(project.URL, gitlabProject)
		return gitlabProject, nil
	}
	return loadedProject.(*gitlab.Project), nil
}

func (b *GitLabProject) mergerequests(project Project) string {
	if !isGitLab(project) {
		return ""
	}

	state := "opened"
	options := &gitlab.ListProjectMergeRequestsOptions{
		State: &state,
	}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	_, response, err := b.client.MergeRequests.ListProjectMergeRequests(id, options)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "mergerequests", "merge requests", "Error", badge.ColorLightgrey, project.URL, err)
	}

	color := badge.ColorBrightgreen
	if response.TotalItems > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "mergerequests", "merge requests", fmt.Sprintf("%d", response.TotalItems), color, project.URL, nil)
}

func (b *GitLabProject) branches(project Project) string {
	if !isGitLab(project) {
		return ""
	}

	options := &gitlab.ListBranchesOptions{}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	_, response, err := b.client.Branches.ListBranches(id, options)
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
	return svgBadge(project.Hoster, project.Name, "branches", "branches", fmt.Sprintf("%d", response.TotalItems), color, project.URL, nil)
}

func (b *GitLabProject) tag(project Project) string {
	if !isGitLab(project) {
		return ""
	}

	options := &gitlab.ListTagsOptions{}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	tags, _, err := b.client.Tags.ListTags(id, options)
	if err != nil {
		return svgBadge(project.Hoster, project.Name, "tag", "tag", "Error", badge.ColorLightgrey, project.URL, err)
	}
	if len(tags) == 0 {
		return ""
	}
	tag := tags[0]
	return svgBadge(project.Hoster, project.Name, "tag", "tag", tag.Name, badge.ColorBlue, project.URL, nil)
}

func (b *GitLabProject) issues(project Project) string {
	gitlabProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}

	color := badge.ColorBrightgreen
	if gitlabProject.OpenIssuesCount > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Hoster, project.Name, "issues", "issues", fmt.Sprintf("%d", gitlabProject.OpenIssuesCount), color, project.URL, nil)
}

func (b *GitLabProject) lastcommit(project Project) string {
	gitlabProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
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
	return svgBadge(project.Hoster, project.Name, "last", "last commit", humanize.Time(*gitlabProject.LastActivityAt), color, project.URL, nil)
}

func (b *GitLabProject) stars(project Project) string {
	gitlabProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "stars", "stars", fmt.Sprint(gitlabProject.StarCount), badge.ColorBlue, project.URL, nil)
}

func (b *GitLabProject) visibility(project Project) string {
	gitlabProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
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

func (b *GitLabProject) forks(project Project) string {
	gitlabProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "fork", "Fork", fmt.Sprint(gitlabProject.ForksCount), badge.ColorBlue, project.URL, nil)
}

func (b *GitLabProject) size(project Project) string {
	gitlabProject, errBadge := b.getProject(project)
	if errBadge != nil {
		return *errBadge
	}
	return svgBadge(project.Hoster, project.Name, "reposize", "repo size", humanize.Bytes(uint64(gitlabProject.Statistics.RepositorySize)), badge.ColorBlue, project.URL, nil)
}
