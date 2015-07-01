package stash

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDeleteBranchRestriction(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("wanted DELETE but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/branch-permissions/1.0/projects/PROJ/repos/slug/restricted/1" {
			t.Fatalf("DeleteBranchRestrictions() URL path expected to be /rest/branch-permissions/1.0/projects/PROJ/repos/slug/restricted/1 but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("DeleteBranchRestrictions() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		w.WriteHeader(204)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	err := stashClient.DeleteBranchRestriction("PROJ", "slug", 1)
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}
}

func TestDeleteBranchRestriction404(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	err := stashClient.DeleteBranchRestriction("PROJ", "slug", 1)
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}

func TestDeleteBranchRestrilction401(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	err := stashClient.DeleteBranchRestriction("PROJ", "slug", 1)
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}
