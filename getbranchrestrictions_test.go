package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const branchRestrictionsResponse string = `
{
  "size": 1,
  "limit": 100,
  "isLastPage": true,
  "values": [
    {
      "id": 41,
      "type": "BRANCH",
      "value": "refs/heads/develop",
      "branch": {
        "id": "refs/heads/develop",
        "displayId": "develop",
        "latestChangeset": "d81c71b179c08715eb21251824635ce9a1d7f6f3",
        "isDefault": false
      }
    }
  ],
  "start": 0,
  "filter": null
}
`

func TestGetBranchRestrictions(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/branch-permissions/1.0/projects/PROJ/repos/slug/restricted" {
			t.Fatalf("GetBranchPermissions() URL path expected to be /rest/branch-permissions/1.0/projects/PROJ/repos/slug/restricted but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetBranchRestrictions() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, branchRestrictionsResponse)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	branchRestrictions, err := stashClient.GetBranchRestrictions("PROJ", "slug")
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}

	// spot checks
	if branchRestrictions.BranchRestriction[0].Branch.DisplayID != "develop" {
		t.Fatalf("Want develop but got %s\n", branchRestrictions.BranchRestriction[0].Branch.DisplayID)
	}
	if branchRestrictions.BranchRestriction[0].Id != 41 {
		t.Fatalf("Want 41 but got %s\n", branchRestrictions.BranchRestriction[0].Id)
	}
}

func TestGetBranchRestrictions404(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	_, err := stashClient.GetBranchRestrictions("PROJ", "slug")
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}

func TestGetBranchRestrictions401(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	_, err := stashClient.GetBranchRestrictions("PROJ", "slug")
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}
