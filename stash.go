// Atlassian Stash API package.
// Stash API Reference: https://developer.atlassian.com/static/rest/stash/3.0.1/stash-rest.html
package stash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"time"
)

type (
	Stash interface {
		GetRepositories() (map[int]Repository, error)
		GetBranches(projectKey, repositorySlug string) (map[string]Branch, error)
		GetRepository(projectKey, repositorySlug string) (Repository, error)
	}

	Client struct {
		userName string
		password string
		baseURL  string
		Stash
	}

	Repositories struct {
		IsLastPage    bool         `json:"isLastPage"`
		Size          int          `json:"size"`
		Start         int          `json:"start"`
		NextPageStart int          `json:"nextPageStart"`
		Repository    []Repository `json:"values"`
	}

	Repository struct {
		ID      int     `json:"id"`
		Name    string  `json:"name"`
		Slug    string  `json:"slug"`
		Project Project `json:"project"`
		ScmID   string  `json:"scmId"`
		Links   Links   `json:"links"`
	}

	Project struct {
		Key string `json:"key"`
	}

	Links struct {
		Clones []Clone `json:"clone"`
	}

	Clone struct {
		HREF string `json:"href"`
		Name string `json:"name"`
	}

	Branches struct {
		IsLastPage    bool     `json:"isLastPage"`
		Size          int      `json:"size"`
		Start         int      `json:"start"`
		NextPageStart int      `json:"nextPageStart"`
		Branch        []Branch `json:"values"`
	}

	Branch struct {
		ID              string `json:"id"`
		DisplayID       string `json:"displayId"`
		LatestChangeSet string `json:"latestChangeset"`
		IsDefault       bool   `json:"isDefault"`
	}
)

const (
	stashPageLimit int = 25
)

var (
	httpClient *http.Client = &http.Client{Timeout: 10 * time.Second}
)

func NewClient(userName, password, baseURL string) Stash {
	return Client{userName: userName, password: password, baseURL: baseURL}
}

// GetRepositories returns a map of repositories indexed by repository URL.
func (client Client) GetRepositories() (map[int]Repository, error) {
	start := 0
	repositories := make(map[int]Repository)
	morePages := true
	for morePages {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/repos?start=%d&limit=%d", client.baseURL, start, stashPageLimit), nil)
		if err != nil {
			return nil, err
		}
		log.Printf("stash.GetRepositories URL %s\n", req.URL)
		req.Header.Set("Accept", "application/json")

		responseCode, data, err := consumeResponse(req)
		if err != nil {
			return nil, err
		}
		if responseCode != http.StatusOK {
			var reason string = "unhandled reason"
			switch {
			case responseCode == http.StatusBadRequest:
				reason = "Bad request."
			}
			return nil, fmt.Errorf("Error getting repositories: %s.  Status code: %d.  Reason: %s\n", string(data), responseCode, reason)
		}

		var r Repositories
		err = json.Unmarshal(data, &r)
		if err != nil {
			return nil, err
		}

		for _, repo := range r.Repository {
			repositories[repo.ID] = repo
		}

		morePages = !r.IsLastPage
		start = r.NextPageStart
	}
	return repositories, nil
}

// GetBranches returns a map of branches indexed by branch display name for the given repository.
func (client Client) GetBranches(projectKey, repositorySlug string) (map[string]Branch, error) {
	start := 0
	branches := make(map[string]Branch)
	morePages := true
	for morePages {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/branches?start=%d&limit=%d", client.baseURL, projectKey, repositorySlug, start, stashPageLimit), nil)
		if err != nil {
			return nil, err
		}
		log.Printf("stash.GetBranches URL %s\n", req.URL)
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(client.userName, client.password)

		responseCode, data, err := consumeResponse(req)
		if err != nil {
			return nil, err
		}

		if responseCode != http.StatusOK {
			var reason string = "unhandled reason"
			switch {
			case responseCode == http.StatusNotFound:
				reason = "Not found"
			case responseCode == http.StatusUnauthorized:
				reason = "Unauthorized"
			}
			return nil, fmt.Errorf("Error getting repository branches: %s.  Status code: %d.  Reason: %s\n", string(data), responseCode, reason)
		}

		var r Branches
		err = json.Unmarshal(data, &r)
		if err != nil {
			return nil, err
		}

		for _, branch := range r.Branch {
			branches[branch.DisplayID] = branch
		}

		morePages = !r.IsLastPage
		start = r.NextPageStart
	}
	return branches, nil
}

// GetRepository returns a repository representation for the given Stash Project key and repository slug.
func (client Client) GetRepository(projectKey, repositorySlug string) (Repository, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s", client.baseURL, projectKey, repositorySlug), nil)
	if err != nil {
		return Repository{}, err
	}
	log.Printf("stash.GetRepository %s\n", req.URL)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept", "application/json")

	responseCode, data, err := consumeResponse(req)
	if err != nil {
		return Repository{}, err
	}

	if responseCode != http.StatusOK {
		var reason string = "unhandled reason"
		switch {
		case responseCode == http.StatusNotFound:
			reason = "Not found"
		case responseCode == http.StatusUnauthorized:
			reason = "Unauthorized"
		}
		return Repository{}, fmt.Errorf("Error getting repository: %s.  Status code: %d.  Reason: %s\n", string(data), responseCode, reason)
	}

	var r Repository
	err = json.Unmarshal(data, &r)
	if err != nil {
		return Repository{}, err
	}

	return r, nil
}

func HasRepository(repositories map[int]Repository, url string) (Repository, bool) {
	for _, repo := range repositories {
		for _, clone := range repo.Links.Clones {
			if clone.HREF == url {
				return repo, true
			}
		}
	}
	return Repository{}, false
}

func consumeResponse(req *http.Request) (rc int, buffer []byte, err error) {
	response, err := httpClient.Do(req)

	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
		if e := recover(); e != nil {
			trace := make([]byte, 10*1024)
			_ = runtime.Stack(trace, false)
			log.Printf("%s", trace)
			err = fmt.Errorf("%v", e)
		}
	}()

	if err != nil {
		panic(err)
	}

	if data, err := ioutil.ReadAll(response.Body); err != nil {
		panic(err)
	} else {
		return response.StatusCode, data, nil
	}
}

// SshUrl extracts the SSH-based URL from the repository metadata.
func (repo Repository) SshUrl() string {
	for _, clone := range repo.Links.Clones {
		if clone.Name == "ssh" {
			return clone.HREF
		}
	}
	return ""
}
