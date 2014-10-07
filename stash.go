// Atlassian Stash API package.
// Stash API Reference: https://developer.atlassian.com/static/rest/stash/3.0.1/stash-rest.html
package stash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	stashPageLimit int = 25
)

var (
	httpClient *http.Client = &http.Client{}
)

// GetRepositories returns a map of repositories indexed by repository URL.
func GetRepositories(baseUrl string) (map[int]Repository, error) {
	start := 0
	repositories := make(map[int]Repository)
	morePages := true
	for morePages {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/repos?start=%d&limit=%d", baseUrl, start, stashPageLimit), nil)
		if err != nil {
			return nil, err
		}
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
func GetBranches(baseUrl, userName, password, projectKey, repositorySlug string) (map[string]Branch, error) {
	start := 0
	branches := make(map[string]Branch)
	morePages := true
	for morePages {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/branches?start=%d&limit=%d", baseUrl, projectKey, repositorySlug, start, stashPageLimit), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.SetBasicAuth(userName, password)

		responseCode, data, err := consumeResponse(req)
		if responseCode != http.StatusOK {
			var reason string = "unhandled reason"
			switch {
			case responseCode == http.StatusNotFound:
				reason = "Not found"
			case responseCode == http.StatusUnauthorized:
				reason = "Unauthorized"
			}
			return nil, fmt.Errorf("Error creating repository: %s.  Status code: %d.  Reason: %s\n", string(data), responseCode, reason)
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

func consumeResponse(req *http.Request) (int, []byte, error) {
	var response *http.Response
	var err error
	response, err = httpClient.Do(req)

	if err != nil {
		return 0, nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	return response.StatusCode, data, nil
}
