package stash

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetRawFile(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("wanted GET but found %s\n", r.Method)
		}
		url := *r.URL

		wantPath := "/projects/prj/repos/repo/browse/foo/bar"
		if url.Path != wantPath {
			t.Fatalf("Want %s but found %s\n", wantPath, url.Path)
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want  Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		params := url.Query()
		if params.Get("at") != "master" {
			t.Fatalf("Want master but found %s\n", params["at"])
		}
		if _, ok := params["raw"]; !ok {
			t.Fatalf("Want a raw query param but found none")
		}

		fmt.Fprint(w, "hello")
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	data, _ := stashClient.GetRawFile("PRJ", "REPO", "foo/bar", "master")
	if string(data) != "hello" {
		t.Fatalf("Want hello, but got <%s>\n", string(data))
	}
}
