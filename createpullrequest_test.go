package stash

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const createPullRequestResponse string = `
{
    "id": 2,
    "version": 0,
    "title": "a title",
    "description": "a description",
    "state": "OPEN",
    "open": true,
    "closed": false,
    "createdDate": 1435759062673,
    "updatedDate": 1435759062673,
    "fromRef": {
        "id": "refs/heads/feature/file1",
        "displayId": "feature/file1",
        "latestChangeset": "aead30bdfe27e176316bb2e2aedd530052730092",
        "repository": {
            "slug": "test-repo",
            "id": 1419,
            "name": "test-repo",
            "scmId": "git",
            "state": "AVAILABLE",
            "statusMessage": "Available",
            "forkable": true,
            "project": {
                "key": "PLAT",
                "id": 349,
                "name": "plat",
                "description": "plat Bill Pay and Top Up Microservices",
                "public": true,
                "type": "NORMAL",
                "isPersonal": false,
                "link": {
                    "url": "/projects/plat",
                    "rel": "self"
                },
                "links": {
                    "self": [
                        {
                            "href": "http://localhost:7990/projects/plat"
                        }
                    ]
                }
            },
            "public": false,
            "link": {
                "url": "/projects/plat/repos/test-repo/browse",
                "rel": "self"
            },
            "cloneUrl": "http://admin@localhost:7990/scm/stamp/test-repo.git",
            "links": {
                "clone": [
                    {
                        "href": "http://admin@localhost:7990/scm/stamp/test-repo.git",
                        "name": "http"
                    },
                    {
                        "href": "ssh://git@localhost:7999/stamp/test-repo.git",
                        "name": "ssh"
                    }
                ],
                "self": [
                    {
                        "href": "http://localhost:7990/projects/plat/repos/test-repo/browse"
                    }
                ]
            }
        }
    },
    "toRef": {
        "id": "refs/heads/develop",
        "displayId": "develop",
        "latestChangeset": "3558d035edf10cb54e316374b9e8403a686995ac",
        "repository": {
            "slug": "test-repo",
            "id": 1419,
            "name": "test-repo",
            "scmId": "git",
            "state": "AVAILABLE",
            "statusMessage": "Available",
            "forkable": true,
            "project": {
                "key": "PLAT",
                "id": 349,
                "name": "Platform Dev",
                "description": "It's a project",
                "public": true,
                "type": "NORMAL",
                "isPersonal": false,
                "link": {
                    "url": "/projects/plat",
                    "rel": "self"
                },
                "links": {
                    "self": [
                        {
                            "href": "http://localhost:7990/projects/plat"
                        }
                    ]
                }
            },
            "public": false,
            "link": {
                "url": "/projects/plat/repos/test-repo/browse",
                "rel": "self"
            },
            "cloneUrl": "http://admin@localhost:7990/scm/stamp/test-repo.git",
            "links": {
                "clone": [
                    {
                        "href": "http://admin@localhost:7990/scm/stamp/test-repo.git",
                        "name": "http"
                    },
                    {
                        "href": "ssh://git@localhost:7999/stamp/test-repo.git",
                        "name": "ssh"
                    }
                ],
                "self": [
                    {
                        "href": "http://localhost:7990/projects/plat/repos/test-repo/browse"
                    }
                ]
            }
        }
    },
    "author": {
        "user": {
            "name": "mike",
            "emailAddress": "mike@myemail.com",
            "id": 877,
            "displayName": "Mike",
            "active": true,
            "slug": "mike",
            "type": "NORMAL",
            "link": {
                "url": "/users/mike",
                "rel": "self"
            },
            "links": {
                "self": [
                    {
                        "href": "http://localhost:7990/users/mike"
                    }
                ]
            }
        },
        "role": "AUTHOR",
        "approved": false
    },
    "reviewers": [
        {
            "user": {
                "name": "bob",
                "emailAddress": "bob@myemail.com",
                "id": 871,
                "displayName": "Bob",
                "active": true,
                "slug": "bob",
                "type": "NORMAL",
                "link": {
                    "url": "/users/bob",
                    "rel": "self"
                },
                "links": {
                    "self": [
                        {
                            "href": "http://localhost:7990/users/bob"
                        }
                    ]
                }
            },
            "role": "REVIEWER",
            "approved": false
        },
        {
            "user": {
                "name": "bill",
                "emailAddress": "bill@myemail.com",
                "id": 871,
                "displayName": "Bill",
                "active": true,
                "slug": "bill",
                "type": "NORMAL",
                "link": {
                    "url": "/users/bill",
                    "rel": "self"
                },
                "links": {
                    "self": [
                        {
                            "href": "http://localhost:7990/users/bill"
                        }
                    ]
                }
            },
            "role": "REVIEWER",
            "approved": false
        }
    ],
    "participants": [],
    "link": {
        "url": "/projects/plat/repos/test-repo/pull-requests/2",
        "rel": "self"
    },
    "links": {
        "self": [
            {
                "href": "http://localhost:7990/projects/plat/repos/test-repo/pull-requests/2"
            }
        ]
    }
}
`

type PullRequestTemplate struct {
	Title       string
	Description string
}

func TestCreatePullRequest(t *testing.T) {

	expectedRequestBody := `{"title":"a title","description":"a description","fromRef":{"id":"feature/file1","repository":{"slug":"bar","project":{"key":"proj"}}},"toRef":{"id":"develop","repository":{"slug":"bar","project":{"key":"proj"}}},"reviewers":[{"user":{"name":"bob"}},{"user":{"name":"bill"}}]}`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("wanted POST but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/projects/proj/repos/bar/pull-requests" {
			t.Fatalf("CreateRepository() URL path expected to be /rest/api/1.0/projects/proj/repos/bar/pull-requests but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("CreateRepository() expected request Accept header to be application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}

		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != string(expectedRequestBody) {
			t.Fatalf("Unexpected request body\n %s\n expected\n %s\n", body, expectedRequestBody)
		}

		w.WriteHeader(201)
		fmt.Fprint(w, createPullRequestResponse)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)

	reviewers := []string{"bob", "bill"}
	pullRequest, err := stashClient.CreatePullRequest("proj", "bar", "a title", "a description", "feature/file1", "develop", reviewers)
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	expect_to_equal(t, "ID", 2, pullRequest.ID)
	expect_to_equal(t, "Title", "a title", pullRequest.Title)
	expect_to_equal(t, "Description", "a description", pullRequest.Description)
	expect_to_equal(t, "Open", true, pullRequest.Open)
	expect_to_equal(t, "State", "OPEN", pullRequest.State)
	expect_to_equal(t, "FromRef", Ref{"feature/file1"}, pullRequest.FromRef)
	expect_to_equal(t, "ToRef", Ref{"develop"}, pullRequest.ToRef)

}

func expect_to_equal(t *testing.T, item string, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Fatalf("expected %s to be <%T>%q got <%T>%q\n", item, actual, actual, expected, expected)
	}
}
