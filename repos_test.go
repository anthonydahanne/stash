package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	repos string = `
{
   "nextPageStart" : 44,
   "isLastPage" : true,
   "filter" : null,
   "limit" : 25,
   "values" : [
      {
         "link" : {
            "rel" : "self",
            "url" : "/projects/TEAMP/repos/apa/browse"
         },
         "project" : {
            "link" : {
               "rel" : "self",
               "url" : "/projects/TEAMP"
            },
            "name" : "Dev",
            "isPersonal" : false,
            "description" : "The development team's stash.",
            "key" : "TEAMP",
            "public" : true,
            "id" : 107,
            "type" : "NORMAL",
            "links" : {
               "self" : [
                  {
                     "href" : "http://example.com:8888/projects/TEAMP"
                  }
               ]
            }
         },
         "name" : "apa",
         "state" : "AVAILABLE",
         "scmId" : "git",
         "cloneUrl" : "http://example.com:8888/scm/teamp/apa.git",
         "statusMessage" : "Available",
         "slug" : "apa",
         "public" : false,
         "forkable" : true,
         "id" : 300,
         "links" : {
            "clone" : [
               {
                  "href" : "ssh://git@example.com:9999/teamp/apa.git",
                  "name" : "ssh"
               },
               {
                  "href" : "http://example.com:8888/scm/teamp/apa.git",
                  "name" : "http"
               }
            ],
            "self" : [
               {
                  "href" : "http://example.com:8888/projects/TEAMP/repos/apa/browse"
               }
            ]
         }
      },
      {
         "link" : {
            "rel" : "self",
            "url" : "/projects/PSC/repos/apac/browse"
         },
         "project" : {
            "link" : {
               "rel" : "self",
               "url" : "/projects/PSC"
            },
            "name" : "Service Clients",
            "isPersonal" : false,
            "description" : "Clients used by dev to call APIs.",
            "key" : "PSC",
            "public" : true,
            "id" : 315,
            "type" : "NORMAL",
            "links" : {
               "self" : [
                  {
                     "href" : "http://example.com:8888/projects/PSC"
                  }
               ]
            }
         },
         "name" : "apac",
         "state" : "AVAILABLE",
         "scmId" : "git",
         "cloneUrl" : "http://example.com:8888/scm/psc/apac.git",
         "statusMessage" : "Available",
         "slug" : "apac",
         "public" : false,
         "forkable" : true,
         "id" : 359,
         "links" : {
            "clone" : [
               {
                  "href" : "ssh://git@example.com:9999/psc/apac.git",
                  "name" : "ssh"
               },
               {
                  "href" : "http://example.com:8888/scm/psc/apac.git",
                  "name" : "http"
               }
            ],
            "self" : [
               {
                  "href" : "http://example.com:8888/projects/PSC/repos/apac/browse"
               }
            ]
         }
      },
      {
         "link" : {
            "rel" : "self",
            "url" : "/projects/TEAMI/repos/rabbit/browse"
         },
         "project" : {
            "link" : {
               "rel" : "self",
               "url" : "/projects/TEAMI"
            },
            "name" : "Infrastructure",
            "isPersonal" : false,
            "description" : "Infrastructure related projects",
            "key" : "TEAMI",
            "public" : true,
            "id" : 143,
            "type" : "NORMAL",
            "links" : {
               "self" : [
                  {
                     "href" : "http://example.com:8888/projects/TEAMI"
                  }
               ]
            }
         },
         "name" : "rabbit",
         "state" : "AVAILABLE",
         "scmId" : "git",
         "cloneUrl" : "http://example.com:8888/scm/teami/rabbit.git",
         "statusMessage" : "Available",
         "slug" : "rabbit",
         "public" : false,
         "forkable" : true,
         "id" : 171,
         "links" : {
            "clone" : [
               {
                  "href" : "http://example.com:8888/scm/teami/rabbit.git",
                  "name" : "http"
               },
               {
                  "href" : "ssh://git@example.com:9999/teami/rabbit.git",
                  "name" : "ssh"
               }
            ],
            "self" : [
               {
                  "href" : "http://example.com:8888/projects/TEAMI/repos/rabbit/browse"
               }
            ]
         }
      }
   ],
   "size" : 25,
   "start" : 0
}
`
)

func TestGetRepositories(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/repos" {
			t.Fatalf("GetRepositories() URL path expected to be /rest/api/1.0/repos but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetRepositories() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, repos)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	repositories, err := stashClient.GetRepositories()
	if err != nil {
		t.Fatalf("GetRepositories() not expecting an error, but received: %v\n", err)
	}

	if len(repositories) != 3 {
		t.Fatalf("GetRepositories() expected to return map of size 3, but received map of size %d\n", len(repositories))
	}

	for _, i := range []int{171, 300, 359} {
		if _, ok := repositories[i]; !ok {
			t.Fatalf("Wanted a repository with ID==%d but found none\n", i)
		}
	}

	for id, slug := range map[int]string{
		171: "rabbit",
		300: "apa",
		359: "apac",
	} {
		if repositories[id].Slug != slug {
			t.Fatalf("Wanted slug==%s for key %d but found %s\n", slug, id, repositories[id].Slug)
		}
	}

}

func TestGetRepositories500(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/repos" {
			t.Fatalf("GetRepositories() URL path expected to be /rest/api/1.0/repos but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("GetRepositories() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		w.WriteHeader(500)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	if _, err := stashClient.GetRepositories(); err == nil {
		t.Fatalf("GetRepositories() expecting an error, but received none\n")
	}
}
