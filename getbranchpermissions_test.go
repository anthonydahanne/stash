package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const branchPermissionsResponse string = `
{
   "type" : "type",
   "matcherType" : "matcherType",
   "matcherId" : "matcherId",
   "effective" : true
}
`

func TestGetBranchPermissions(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/branch-permissions/1.0/projects/PROJ/repos/slug/permitted" {
			t.Fatalf("GetBranchPermissions() URL path expected to be /rest/branch-permissions/1.0/projects/PROJ/repos/slug/permitted but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetBranchPermissions() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, branchPermissionsResponse)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	branchPermissions, err := stashClient.GetBranchPermissions("PROJ", "slug")
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}

	// spot checks
	if branchPermissions.Type != "type" {
		t.Fatalf("Want type but got %s\n", branchPermissions.Type)
	}
	if branchPermissions.MatcherType != "matcherType" {
		t.Fatalf("Want matcherType but got %s\n", branchPermissions.MatcherType)
	}
	if branchPermissions.MatcherId != "matcherId" {
		t.Fatalf("Want matcherId but got %s\n", branchPermissions.MatcherId)
	}
	if !branchPermissions.Effective {
		t.Fatalf("Want true but got %s\n", branchPermissions.Effective)
	}
}

//func TestGetBranchPermissions404(t *testing.T) {
//	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(404)
//	}))
//	defer testServer.Close()
//
//	url, _ := url.Parse(testServer.URL)
//	stashClient := NewClient("u", "p", url)
//	_, err := stashClient.GetRepository("PROJ", "slug")
//	if err == nil {
//		t.Fatalf("Expecting error but did not get one\n")
//	}
//}
//
//func TestGetBranchPermissions401(t *testing.T) {
//	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(401)
//	}))
//	defer testServer.Close()
//
//	url, _ := url.Parse(testServer.URL)
//	stashClient := NewClient("u", "p", url)
//	_, err := stashClient.GetRepository("PROJ", "slug")
//	if err == nil {
//		t.Fatalf("Expecting error but did not get one\n")
//	}
//}
