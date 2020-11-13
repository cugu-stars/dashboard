package badge

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/narqo/go-badge"
)

func InitExternalCommandBadges() {
	badges["pycodestyle"] = pycodestyle
	badges["superlint"] = superlint
	badges["bandit"] = bandit
	badges["shhgit"] = shhgit
}

func shhgit(project Project) *Badge {
	projectPath, err := download(project)
	if err != nil {
		return errorBadge("shhgit", project, err)
	}

	cmd := exec.Command("/root/go/bin/shhgit", "--config-path", "/shhgit", "--local", projectPath)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		shhgitLog := filepath.Join("badges", project.Hoster, project.Name, "shhgit.txt")
		b := svgBadge(project.Hoster, project.Name, "shhgit", "shhgit", "invalid", badge.ColorRed, shhgitLog, nil)
		_ = ioutil.WriteFile(shhgitLog, out.Bytes(), 0666)
		return b
	}
	return svgBadge(project.Hoster, project.Name, "shhgit", "shhgit", "valid", badge.ColorBrightgreen, "https://github.com/eth0izzle/shhgit", nil)
}

func bandit(project Project) *Badge {
	projectPath, err := download(project)
	if err != nil {
		return errorBadge("bandit", project, err)
	}

	cmd := exec.Command("bandit", "-r", projectPath)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		banditLog := filepath.Join("badges", project.Hoster, project.Name, "bandit.txt")
		b := svgBadge(project.Hoster, project.Name, "bandit", "bandit", "invalid", badge.ColorRed, banditLog, nil)
		_ = ioutil.WriteFile(banditLog, out.Bytes(), 0666)
		return b
	}
	return svgBadge(project.Hoster, project.Name, "bandit", "bandit", "valid", badge.ColorBrightgreen, "https://pypi.org/project/bandit/", nil)
}

func pycodestyle(project Project) *Badge {
	projectPath, err := download(project)
	if err != nil {
		return errorBadge("pycodestyle", project, err)
	}

	cmd := exec.Command("pycodestyle", projectPath)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		pycodestyleLog := filepath.Join("badges", project.Hoster, project.Name, "pycodestyle.txt")
		b := svgBadge(project.Hoster, project.Name, "pycodestyle", "pycodestyle", "invalid", badge.ColorRed, pycodestyleLog, nil)
		_ = ioutil.WriteFile(pycodestyleLog, out.Bytes(), 0666)
		return b
	}
	return svgBadge(project.Hoster, project.Name, "pycodestyle", "pycodestyle", "valid", badge.ColorBrightgreen, "https://pypi.org/project/pycodestyle/", nil)
}

func superlint(project Project) *Badge {
	projectPath, err := download(project)
	if err != nil {
		return errorBadge("superlint", project, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel() // The cancel should be deferred so resources are cleaned up
	cmd := exec.CommandContext(ctx, "/action/lib/linter.sh")
	cmd.Env = append(os.Environ(),
		"RUN_LOCAL=true",
		"DEFAULT_WORKSPACE="+projectPath,
		// "LOG_LEVEL=DEBUG",
		// "OUTPUT_FORMAT=tap",
		// "OUTPUT_FOLDER=report",
		// "OUTPUT_DETAILS=detailed",
	)
	var report bytes.Buffer
	cmd.Stderr = &report
	err = cmd.Run()

	_ = os.MkdirAll(filepath.Join("badges", project.Hoster, project.Name), 0777)
	reportData := ansiRe.ReplaceAll(report.Bytes(), []byte{})
	reportData = logRe.ReplaceAll(reportData, []byte{})
	lintLog := filepath.Join("badges", project.Hoster, project.Name, "super-linter.txt")
	_ = ioutil.WriteFile(lintLog, reportData, 0666)

	if err != nil {
		return svgBadge(project.Hoster, project.Name, "super-linter", "super-linter", "invalid", badge.ColorRed, lintLog, nil)
		/*
			infos, err := ioutil.ReadDir(path.Join(projectPath, "report"))
			if err != nil {
				log.Println(err)
			}
			for _, info := range infos {
				b, err := ioutil.ReadFile(path.Join(projectPath, "report", info.Name()))
				if err != nil {
					log.Println(err)
					continue
				}
				report.WriteString("\n\n#" + path.Join(projectPath, "report", info.Name()) + "\n\n")


				_, _ = report.Write(b)
			}
		*/
	}
	return svgBadge(project.Hoster, project.Name, "super-linter", "super-linter", "valid", badge.ColorBrightgreen, lintLog, nil)
}

var ansiRe = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
var logRe = regexp.MustCompile(`.*?\]   `)
