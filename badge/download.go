package badge

import (
	"crypto/tls"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

var downloaded = map[string]string{}
var downloadLock = sync.Mutex{}

func download(project Project) (string, error) {
	downloadLock.Lock()
	defer downloadLock.Unlock()

	if name, ok := downloaded[project.URL]; ok {
		return name, nil
	}

	name, err := ioutil.TempDir("", "git")
	if err != nil {
		return "", err
	}

	customClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: Insecure},
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: 10 * time.Second,
	}
	client.InstallProtocol("https", githttp.NewClient(customClient))

	_, err = git.PlainClone(name, false, &git.CloneOptions{Auth: &githttp.BasicAuth{
		Username: "xx", // yes, this can be anything except an empty string
		Password: project.Token,
	}, URL: project.URL})
	if err != nil {
		return "", err

	}

	downloaded[project.URL] = name
	return name, nil
}
