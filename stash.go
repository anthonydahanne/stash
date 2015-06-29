// Atlassian Stash API package.
// Stash API Reference: https://developer.atlassian.com/static/rest/stash/3.0.1/stash-rest.html
package stash

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/ae6rt/retry"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

type (
	Stash interface {
		CreateRepository(projectKey, slug string) (Repository, error)
		GetRepositories() (map[int]Repository, error)
		GetBranches(projectKey, repositorySlug string) (map[string]Branch, error)
		GetBranchRestrictions(projectKey, repositorySlug string) (BranchRestrictions, error)
		GetRepository(projectKey, repositorySlug string) (Repository, error)
		GetPullRequests(projectKey, repositorySlug, state string) ([]PullRequest, error)
		GetRawFile(projectKey, repositorySlug, branch, filePath string) ([]byte, error)
	}

	Client struct {
		userName string
		password string
		baseURL  *url.URL
		Stash
	}

	Page struct {
		IsLastPage    bool `json:"isLastPage"`
		Size          int  `json:"size"`
		Start         int  `json:"start"`
		NextPageStart int  `json:"nextPageStart"`
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

	BranchRestrictions struct {
		BranchRestriction []BranchRestriction `json:"values"`
	}

	BranchRestriction struct {
		Branch Branch `json:"branch"`
	}

	PullRequests struct {
		Page
		PullRequests []PullRequest `json:"values"`
	}

	PullRequest struct {
		ID          int    `id:"closed"`
		Closed      bool   `json:"closed"`
		Open        bool   `json:"open"`
		State       string `json:"state"`
		Title       string `json:"title"`
		Description string `json:"description"`
		FromRef     Ref    `json:"fromRef"`
		ToRef       Ref    `json:"toRef"`
	}

	Ref struct {
		DisplayID string `json:"displayId"`
	}

	errorResponse struct {
		StatusCode int
		Reason     string
		error
	}
)

const (
	stashPageLimit int = 25
)

var (
	httpTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
)

var (
	httpClient *http.Client = &http.Client{Timeout: 10 * time.Second, Transport: httpTransport}
)

func (e errorResponse) Error() string {
	return fmt.Sprintf("%s (%d)", e.Reason, e.StatusCode)
}

func NewClient(userName, password string, baseURL *url.URL) Stash {
	return Client{userName: userName, password: password, baseURL: baseURL}
}

func (client Client) CreateRepository(projectKey, projectSlug string) (Repository, error) {
	slug := fmt.Sprintf(`{"name": "%s", "scmId": "git"}`, projectSlug)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos", client.baseURL.String(), projectKey), bytes.NewBuffer([]byte(slug)))
	if err != nil {
		return Repository{}, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")
	req.SetBasicAuth(client.userName, client.password)

	responseCode, data, err := consumeResponse(req)
	if err != nil {
		return Repository{}, err
	}
	if responseCode != http.StatusCreated {
		var reason string = "unknown reason"
		switch {
		case responseCode == http.StatusBadRequest:
			reason = "The repository was not created due to a validation error."
		case responseCode == http.StatusUnauthorized:
			reason = "The currently authenticated user has insufficient permissions to create a repository."
		case responseCode == http.StatusNotFound:
			reason = "The resource was not found.  Does the project key exist?"
		case responseCode == http.StatusConflict:
			reason = "A repository with same name already exists."
		}
		return Repository{}, errorResponse{StatusCode: responseCode, Reason: reason}
	}

	var t Repository
	err = json.Unmarshal(data, &t)
	if err != nil {
		return Repository{}, err
	}

	return t, nil
}

// GetRepositories returns a map of repositories indexed by repository URL.
func (client Client) GetRepositories() (map[int]Repository, error) {
	start := 0
	repositories := make(map[int]Repository)
	morePages := true
	for morePages {
		retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)
		var data []byte
		work := func() error {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/repos?start=%d&limit=%d", client.baseURL.String(), start, stashPageLimit), nil)
			if err != nil {
				return err
			}
			log.Printf("stash.GetRepositories URL %s\n", req.URL)
			req.Header.Set("Accept", "application/json")
			req.SetBasicAuth(client.userName, client.password)

			var responseCode int
			responseCode, data, err = consumeResponse(req)
			if err != nil {
				return err
			}
			if responseCode != http.StatusOK {
				var reason string = "unhandled reason"
				switch {
				case responseCode == http.StatusBadRequest:
					reason = "Bad request."
				}
				return errorResponse{StatusCode: responseCode, Reason: reason}
			}
			return nil
		}
		if err := retry.Try(work); err != nil {
			return nil, err
		}

		var r Repositories
		err := json.Unmarshal(data, &r)
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
		var data []byte
		retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)
		workit := func() error {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/branches?start=%d&limit=%d", client.baseURL.String(), projectKey, repositorySlug, start, stashPageLimit), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Accept", "application/json")
			req.SetBasicAuth(client.userName, client.password)

			var responseCode int
			responseCode, data, err = consumeResponse(req)
			if err != nil {
				return err
			}

			if responseCode != http.StatusOK {
				var reason string = "unhandled reason"
				switch {
				case responseCode == http.StatusNotFound:
					reason = "Not found"
				case responseCode == http.StatusUnauthorized:
					reason = "Unauthorized"
				}
				return errorResponse{StatusCode: responseCode, Reason: reason}
			}
			return nil
		}
		if err := retry.Try(workit); err != nil {
			return nil, err
		}

		var r Branches
		if err := json.Unmarshal(data, &r); err != nil {
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
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)

	var r Repository
	work := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s", client.baseURL.String(), projectKey, repositorySlug), nil)
		if err != nil {
			return err
		}
		log.Printf("stash.GetRepository %s\n", req.URL)
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(client.userName, client.password)

		responseCode, data, err := consumeResponse(req)
		if err != nil {
			return err
		}

		if responseCode != http.StatusOK {
			var reason string = "unhandled reason"
			switch {
			case responseCode == http.StatusNotFound:
				reason = "Not found"
			case responseCode == http.StatusUnauthorized:
				reason = "Unauthorized"
			}
			return errorResponse{StatusCode: responseCode, Reason: reason}
		}

		err = json.Unmarshal(data, &r)
		if err != nil {
			return err
		}
		return nil
	}

	return r, retry.Try(work)
}

// GetRepository returns a repository representation for the given Stash Project key and repository slug.
func (client Client) GetBranchRestrictions(projectKey, repositorySlug string) (BranchRestrictions, error) {
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)

	var branchRestrictions BranchRestrictions
	work := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/branch-permissions/1.0/projects/%s/repos/%s/restricted", client.baseURL.String(), projectKey, repositorySlug), nil)
		if err != nil {
			return err
		}
		log.Printf("stash.GetBranchPermissions %s\n", req.URL)
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(client.userName, client.password)

		responseCode, data, err := consumeResponse(req)
		if err != nil {
			return err
		}

		if responseCode != http.StatusOK {
			var reason string = "unhandled reason"
			switch {
			case responseCode == http.StatusNotFound:
				reason = "Not found"
			case responseCode == http.StatusUnauthorized:
				reason = "Unauthorized"
			}
			return errorResponse{StatusCode: responseCode, Reason: reason}
		}

		err = json.Unmarshal(data, &branchRestrictions)
		if err != nil {
			return err
		}
		return nil
	}

	return branchRestrictions, retry.Try(work)
}

// GetPullRequests returns a list of pull requests for a project / slug.
func (client Client) GetPullRequests(projectKey, projectSlug, state string) ([]PullRequest, error) {
	start := 0
	pullRequests := make([]PullRequest, 0)
	morePages := true
	for morePages {
		retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)
		var data []byte
		work := func() error {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/pull-requests?state=%s&start=%d&limit=%d", client.baseURL.String(), projectKey, projectSlug, state, start, stashPageLimit), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Accept", "application/json")
			req.SetBasicAuth(client.userName, client.password)

			var responseCode int
			responseCode, data, err = consumeResponse(req)
			if err != nil {
				return err
			}
			if responseCode != http.StatusOK {
				var reason string = "unhandled reason"
				switch {
				case responseCode == http.StatusBadRequest:
					reason = "Bad request."
				}
				return errorResponse{StatusCode: responseCode, Reason: reason}
			}
			return nil
		}
		if err := retry.Try(work); err != nil {
			return nil, err
		}

		var r PullRequests
		err := json.Unmarshal(data, &r)
		if err != nil {
			return nil, err
		}
		for _, pr := range r.PullRequests {
			pullRequests = append(pullRequests, pr)
		}
		morePages = !r.IsLastPage
		start = r.NextPageStart
	}
	return pullRequests, nil
}

func (client Client) GetRawFile(repositoryProjectKey, repositorySlug, filePath, branch string) ([]byte, error) {
	retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)

	var data []byte
	work := func() error {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/repos/%s/browse/%s?at=%s&raw", client.baseURL.String(), strings.ToLower(repositoryProjectKey), strings.ToLower(repositorySlug), filePath, branch), nil)
		if err != nil {
			return err
		}
		log.Printf("stash.GetRawFile %s\n", req.URL)
		req.SetBasicAuth(client.userName, client.password)

		var responseCode int
		responseCode, data, err = consumeResponse(req)
		if err != nil {
			return err
		}
		if responseCode != http.StatusOK {
			var reason string = "unhandled reason"
			switch {
			case responseCode == http.StatusNotFound:
				reason = "Not found"
			case responseCode == http.StatusUnauthorized:
				reason = "Unauthorized"
			}
			return errorResponse{StatusCode: responseCode, Reason: reason}
		}
		return nil
	}

	return data, retry.Try(work)
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

func IsRepositoryExists(err error) bool {
	if err == nil {
		return false
	}
	if response, ok := err.(errorResponse); ok {
		return response.StatusCode == http.StatusConflict
	}
	return false
}

func IsRepositoryNotFound(err error) bool {
	if err == nil {
		return false
	}
	if response, ok := err.(errorResponse); ok {
		return response.StatusCode == http.StatusNotFound
	}
	return false
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
