package badge

import (
	"os"
	"path"

	"github.com/narqo/go-badge"
)

func InitMissingFileBadges() {
	badges["readme"] = readme
	badges["gitignore"] = gitignore
}

func readme(project Project) *Badge {
	projectPath, err := download(project)
	if err != nil {
		return errorBadge("readme", project, err)
	}

	_, err = os.Stat(path.Join(projectPath, "README"))
	_, errmd := os.Stat(path.Join(projectPath, "README.md"))
	if os.IsNotExist(err) && os.IsNotExist(errmd) {
		return svgBadge(project.Hoster, project.Name, "readme", "Readme", "missing", badge.ColorRed, project.URL, nil)
	}

	return svgBadge(project.Hoster, project.Name, "readme", "Readme", "exists", badge.ColorBrightgreen, project.URL, nil)
}

func gitignore(project Project) *Badge {
	projectPath, err := download(project)
	if err != nil {
		return errorBadge("gitignore", project, err)
	}

	_, err = os.Stat(path.Join(projectPath, ".gitignore"))
	if os.IsNotExist(err) {
		return svgBadge(project.Hoster, project.Name, "gitignore", ".gitignore", "missing", badge.ColorRed, project.URL, nil)
	}

	return svgBadge(project.Hoster, project.Name, "gitignore", ".gitignore", "exists", badge.ColorBrightgreen, project.URL, nil)
}
