package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var createResponse = `
{
    "cloneUrl": "http://admin@localhost:7990/scm/plat/bar.git", 
    "forkable": true, 
    "id": 17, 
    "link": {
        "rel": "self", 
        "url": "/projects/PLAT/repos/bar/browse"
    }, 
    "links": {
        "clone": [
            {
                "href": "ssh://git@localhost:7999/plat/bar.git", 
                "name": "ssh"
            }, 
            {
                "href": "http://admin@localhost:7990/scm/plat/bar.git", 
                "name": "http"
            }
        ], 
        "self": [
            {
                "href": "http://localhost:7990/projects/PLAT/repos/bar/browse"
            }
        ]
    }, 
    "name": "bar", 
    "project": {
        "id": 2, 
        "key": "PLAT", 
        "link": {
            "rel": "self", 
            "url": "/projects/PLAT"
        }, 
        "links": {
            "self": [
                {
                    "href": "http://localhost:7990/projects/PLAT"
                }
            ]
        }, 
        "name": "Platform Dev", 
        "public": false, 
        "type": "NORMAL"
    }, 
    "public": false, 
    "scmId": "git", 
    "slug": "bar", 
    "state": "AVAILABLE", 
    "statusMessage": "Available"
}
`

func TestCreateRepository(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/projects/proj/repos" {
			t.Fatalf("CreateRepository() URL path expected to be /rest/api/1.0/projects/proj/repos but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("CreateRepository() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		w.WriteHeader(201)
		fmt.Fprint(w, createResponse)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	repo, err := stashClient.CreateRepository("proj", "bar")
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}

	// spot checks
	if repo.Slug != "bar" {
		t.Fatalf("Want bar but got %s\n", repo.Slug)
	}
	if repo.ID != 17 {
		t.Fatalf("Want 17 but got %s\n", repo.ID)
	}
	if url := repo.SshUrl(); url != "ssh://git@localhost:7999/plat/bar.git" {
		t.Fatalf("Want ssh://git@localhost:7999/plat/bar.git but got %s\n", repo.ID)
	}
}
