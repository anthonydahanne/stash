package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	branches string = `
{
   "isLastPage" : true,
   "filter" : null,
   "values" : [
      {
         "displayId" : "develop",
         "isDefault" : true,
         "latestChangeset" : "e680a10f3e0afb5e3a5978dea02d37ac884da21",
         "id" : "refs/heads/develop"
      },
      {
         "displayId" : "master",
         "isDefault" : false,
         "latestChangeset" : "8d0f23745dfe4bacef9509bb4ecd7722b9aff82",
         "id" : "refs/heads/master"
      },
      {
         "displayId" : "feature/PRJ-447",
         "isDefault" : false,
         "latestChangeset" : "8d9c0642da6b3f06629cf115683da105d8e0654",
         "id" : "refs/heads/feature/PRJ-447"
      },
      {
         "displayId" : "bug/PRJ-442",
         "isDefault" : false,
         "latestChangeset" : "a57a403996161f24d1d0605ea8b5030927a0d3d",
         "id" : "refs/heads/bug/PRJ-442"
      }
   ],
   "limit" : 25,
   "start" : 0,
   "size" : 7
}
`
)

func TestGetBranches(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/projects/PRJ/repos/widge/branches" {
			t.Fatalf("GetBranches() URL path expected to be /rest/api/1.0/projects/PRJ/repos/widge/branches but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetBranches() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, branches)
	}))
	defer testServer.Close()

	stashClient := NewClient("u", "p", testServer.URL)
	branches, err := stashClient.GetBranches("PRJ", "widge")
	if err != nil {
		t.Fatalf("GetBranches() not expecting an error, but received: %v\n", err)
	}

	if len(branches) != 4 {
		t.Fatalf("GetBranches() expected to return map of size 4, but received map of size %d\n", len(branches))
	}

	for _, i := range []string{"master", "develop", "feature/PRJ-447", "bug/PRJ-442"} {
		if _, ok := branches[i]; !ok {
			t.Fatalf("Wanted a branch with displayID==%s but found none\n", i)
		}
	}
}

func TestGetBranches500(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/projects/PRJ/repos/widge/branches" {
			t.Fatalf("GetBranches() URL path expected to be /rest/api/1.0/projects/PRJ/repos/widge/branches but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetBranches() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		w.WriteHeader(500)
	}))
	defer testServer.Close()

	stashClient := NewClient("u", "p", testServer.URL)
	if _, err := stashClient.GetBranches("PRJ", "widge"); err == nil {
		t.Fatalf("GetBranches() expecting an error but received none\n")
	}
}
