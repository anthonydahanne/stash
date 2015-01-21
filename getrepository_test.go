package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const response string = `
{
   "link" : {
      "rel" : "self",
      "url" : "/projects/PROJ/repos/trunk/browse"
   },
   "project" : {
      "link" : {
         "rel" : "self",
         "url" : "/projects/PROJ"
      },
      "name" : "PROJ Dev",
      "isPersonal" : false,
      "description" : "The PROJ stash.",
      "key" : "PROJ",
      "public" : true,
      "id" : 107,
      "type" : "NORMAL",
      "links" : {
         "self" : [
            {
               "href" : "http://example.com:8888/projects/PROJ"
            }
         ]
      }
   },
   "name" : "trunk",
   "state" : "AVAILABLE",
   "scmId" : "git",
   "cloneUrl" : "http://user@example.com:8888/scm/PROJ/trunk.git",
   "statusMessage" : "Available",
   "public" : false,
   "slug" : "trunk",
   "id" : 536,
   "forkable" : true,
   "links" : {
      "clone" : [
         {
            "href" : "ssh://git@example.com:9999/PROJ/trunk.git",
            "name" : "ssh"
         },
         {
            "href" : "http://user@example.com:8888/scm/PROJ/trunk.git",
            "name" : "http"
         }
      ],
      "self" : [
         {
            "href" : "http://example.com:8888/projects/PROJ/repos/trunk/browse"
         }
      ]
   }
}
`

func TestGetRepository(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/projects/PROJ/repos/slug" {
			t.Fatalf("GetBranches() URL path expected to be /rest/api/1.0/projects/PROJ/repos/slug but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetBranches() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		fmt.Fprintln(w, response)
	}))
	defer testServer.Close()

	stashClient := NewClient("u", "p", testServer.URL)
	repo, err := stashClient.GetRepository("PROJ", "slug")
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}

	// spot checks
	if repo.Slug != "trunk" {
		t.Fatalf("Want trunk but got %s\n", repo.Slug)
	}
	if repo.ID != 536 {
		t.Fatalf("Want 536 but got %s\n", repo.ID)
	}
	if url := repo.SshUrl(); url != "ssh://git@example.com:9999/PROJ/trunk.git" {
		t.Fatalf("Want ssh://git@example.com:9999/PROJ/trunk.git but got %s\n", repo.ID)
	}
}

func TestGetRepository404(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer testServer.Close()

	stashClient := NewClient("u", "p", testServer.URL)
	_, err := stashClient.GetRepository("PROJ", "slug")
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}

func TestGetRepository401(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer testServer.Close()

	stashClient := NewClient("u", "p", testServer.URL)
	_, err := stashClient.GetRepository("PROJ", "slug")
	if err == nil {
		t.Fatalf("Expecting error but did not get one\n")
	}
}
