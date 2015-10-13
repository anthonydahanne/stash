package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var (
	tags string = `
{
   "isLastPage" : true,
   "filter" : null,
   "values" : [
        {
            "displayId": "acme-release-99.8", 
            "hash": "c505d5eac54dc0239c610274f0c972845b4d71c3", 
            "id": "refs/tags/acme-release-99.8", 
            "latestChangeset": "fa6618112e8014934dfdfc3337e94f52b6de5708"
        }, 
        {
            "displayId": "acme-release-99.9", 
            "hash": "cd301bcf63344a9c2a4acf88591961ff9a7bc44b", 
            "id": "refs/tags/acme-release-99.9", 
            "latestChangeset": "f0910c480a77b6ccf919fb384ab87f7ab4fd479e"
        }
   ],
   "limit" : 25,
   "start" : 0,
   "size" : 7
}
`
)

func TestGetTags(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/api/1.0/projects/PRJ/repos/widge/tags" {
			t.Fatalf("Want /rest/api/1.0/projects/PRJ/repos/widge/tags but found %s\n", url.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Fatalf("Want application/json but found %s\n", r.Header.Get("Accept"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, tags)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	branches, err := stashClient.GetTags("PRJ", "widge")
	if err != nil {
		t.Fatal(err)
	}

	if len(branches) != 2 {
		t.Fatalf("Want 4 but got %d\n", len(branches))
	}

	for _, i := range []string{"acme-release-99.8", "acme-release-99.9"} {
		if _, ok := branches[i]; !ok {
			t.Fatalf("Wanted a tag with displayID==%s but found none\n", i)
		}
	}
}

func TestGetTagsAnonymous(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("Want no Authorization header but found one: %s\n", r.Header.Get("Authorization"))
		}
		fmt.Fprintln(w, tags)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("", "", url)
	branches, err := stashClient.GetTags("PRJ", "widge")
	if err != nil {
		t.Fatal(err)
	}

	if len(branches) != 2 {
		t.Fatalf("Want 4 but got %d\n", len(branches))
	}

	for _, i := range []string{"acme-release-99.8", "acme-release-99.9"} {
		if _, ok := branches[i]; !ok {
			t.Fatalf("Wanted a tag with displayID==%s but found none\n", i)
		}
	}
}
