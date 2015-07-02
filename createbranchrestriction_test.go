package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const createBranchRestrictionsResponse string = `
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
`

func TestCreateBranchRestriction(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/branch-permissions/1.0/projects/PROJ/repos/slug/restricted" {
			t.Fatalf("CreateBranchPermissions() URL path expected to be /rest/branch-permissions/1.0/projects/PROJ/repos/slug/restricted but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("CreateBranchPermissions() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, createBranchRestrictionsResponse)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	branchRestriction, err := stashClient.CreateBranchRestriction("PROJ", "slug", "develop", "user")
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}

	// spot checks
	if branchRestriction.Branch.DisplayID != "develop" {
		t.Fatalf("Want develop but got %s\n", branchRestriction.Branch.DisplayID)
	}
	if branchRestriction.Id != 41 {
		t.Fatalf("Want 41 but got %s\n", branchRestriction.Id)
	}
}

func TestCreateBranchRestriction404(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	_, err := stashClient.CreateBranchRestriction("PROJ", "slug", "develop", "user")
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}

func TestCreateBranchRestrilction401(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	_, err := stashClient.CreateBranchRestriction("PROJ", "slug", "develop", "user")
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}
