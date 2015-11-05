package stash

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDeleteBranch(t *testing.T) {
	type deleteModel struct {
		Ref    string `json:"name"`
		DryRun bool   `json:"dryRun"`
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Fatalf("wanted DELETE but found %s\n", r.Method)
		}
		url := *r.URL
		if url.Path != "/rest/branch-utils/1.0/projects/PROJ/repos/slug/branches" {
			t.Fatalf("want /rest/branch-utils/1.0/projects/PROJ/repos/slug/branches but got %s\n", url.Path)
		}
		if r.Header.Get("Content-type") != "application/json" {
			t.Fatalf("Want Content-type application/json but found %s\n", r.Header.Get("Content-type"))
		}
		if r.Header.Get("Authorization") != "Basic dTpw" {
			t.Fatalf("Want Basic dTpw but found %s\n", r.Header.Get("Authorization"))
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Unexpected error: %v\n", err)
		}

		var d deleteModel
		if err := json.Unmarshal(data, &d); err != nil {
			t.Fatalf("Unexpected error: %v\n", err)
		}
		if d.Ref != "refs/heads/issue/1" {
			t.Fatalf("Want refs/heads/issue/1 but got %s\n", d.Ref)
		}
		w.WriteHeader(204)
	}))
	defer testServer.Close()

	url, _ := url.Parse(testServer.URL)
	stashClient := NewClient("u", "p", url)
	err := stashClient.DeleteBranch("PROJ", "slug", "issue/1")
	if err != nil {
		t.Fatalf("Not expecting error: %v\n", err)
	}
}
