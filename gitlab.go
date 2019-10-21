package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/narqo/go-badge"
	"github.com/xanzy/go-gitlab"
)

type GitLabMergeRequests struct{}

func (b *GitLabMergeRequests) Render(column string, project Project) string {
	if !isGitLab(project) {
		return ""
	}
	gl := gitlab.NewClient(nil, GITLAB_ACCESS_TOKEN)

	state := "opened"
	options := &gitlab.ListProjectMergeRequestsOptions{
		State: &state,
	}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	_, response, err := gl.MergeRequests.ListProjectMergeRequests(id, options)
	if err != nil {
		return svgBadge(project.Name, "mergerequests", "merge requests", err.Error(), badge.ColorLightgrey, project.URL)
	}

	color := badge.ColorBrightgreen
	if response.TotalItems > 0 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Name, "mergerequests", "merge requests", fmt.Sprintf("%d open", response.TotalItems), color, project.URL)
}

type GitLabBranches struct{}

func (b *GitLabBranches) Render(column string, project Project) string {
	if !isGitLab(project) {
		return ""
	}
	gl := gitlab.NewClient(nil, GITLAB_ACCESS_TOKEN)

	options := &gitlab.ListBranchesOptions{}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	_, response, err := gl.Branches.ListBranches(id, options)
	if err != nil {
		return svgBadge(project.Name, "branches", "branches", err.Error(), badge.ColorLightgrey, project.URL)
	}

	color := badge.ColorBrightgreen
	if response.TotalItems > 1 {
		color = badge.ColorGreen
	}
	if response.TotalItems > 2 {
		color = badge.ColorYellow
	}
	return svgBadge(project.Name, "branches", "branches", fmt.Sprintf("%d open", response.TotalItems), color, project.URL)
}

type GitLabTag struct{}

func (b *GitLabTag) Render(column string, project Project) string {
	if !isGitLab(project) {
		return ""
	}
	gl := gitlab.NewClient(nil, GITLAB_ACCESS_TOKEN)

	options := &gitlab.ListTagsOptions{}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")
	tags, _, err := gl.Tags.ListTags(id, options)
	if err != nil {
		return svgBadge(project.Name, "tag", "tag", err.Error(), badge.ColorLightgrey, project.URL)
	}
	if len(tags) == 0 {
		return ""
	}
	tag := tags[0]
	return svgBadge(project.Name, "tag", "tag", tag.Name, badge.ColorBlue, project.URL)
}

type GitLabProject struct {
	field string
}

func (b *GitLabProject) Render(column string, project Project) string {
	if !isGitLab(project) {
		return ""
	}
	gl := gitlab.NewClient(nil, GITLAB_ACCESS_TOKEN)

	t := true
	options := &gitlab.GetProjectOptions{
		Statistics: &t,
	}

	id := project.Namespace + "/" + project.Name
	id = strings.Trim(id, "/")

	gitlabProject, _, err := gl.Projects.GetProject(id, options)
	if err != nil {
		return svgBadge(project.Name, "gitlab", "gitlab", err.Error(), badge.ColorLightgrey, project.URL)
	}

	switch b.field {
	case "issues":
		color := badge.ColorBrightgreen
		if gitlabProject.OpenIssuesCount > 0 {
			color = badge.ColorYellow
		}
		return svgBadge(project.Name, "issues", "issues", fmt.Sprintf("%d open", gitlabProject.OpenIssuesCount), color, project.URL)
	case "lastcommit":
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
		return svgBadge(project.Name, "last", "last commit", humanize.Time(*gitlabProject.LastActivityAt), color, project.URL)
	case "stars":
		return svgBadge(project.Name, "stars", "stars", fmt.Sprint(gitlabProject.StarCount), badge.ColorBlue, project.URL)
	case "visibility":
		color := badge.ColorBlue
		switch gitlabProject.Visibility {
		case gitlab.PrivateVisibility:
			color = badge.ColorLightgray
		case gitlab.PublicVisibility:
			color = badge.ColorGreen
		}
		return svgBadge(project.Name, "visibility", "visibility", string(gitlabProject.Visibility), color, project.URL)
	case "forks":
		return svgBadge(project.Name, "fork", "Fork", fmt.Sprint(gitlabProject.ForksCount), badge.ColorBlue, project.URL)
	case "size":
		return svgBadge(project.Name, "reposize", "repo size", humanize.Bytes(uint64(gitlabProject.Statistics.RepositorySize)), badge.ColorBlue, project.URL)
	}

	return svgBadge(project.Name, "error", "error", "unknown field", badge.ColorLightgrey, project.URL)
}
