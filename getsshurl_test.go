package stash

import "testing"

func TestGetSshUrl(t *testing.T) {
	clones := []Clone{
		Clone{Name: "ssh", HREF: "ssh-url"},
		Clone{Name: "http", HREF: "http-url"},
	}
	links := Links{Clones: clones}
	repository := Repository{Links: links}
	sshURL := repository.SshUrl()
	if sshURL != "ssh-url" {
		t.Fatalf("Want ssh-url but got %s\n", sshURL)
	}
}

func TestGetSshUrlMissing(t *testing.T) {
	clones := []Clone{
		Clone{Name: "http", HREF: "http-url"},
	}
	links := Links{Clones: clones}
	repository := Repository{Links: links}
	sshURL := repository.SshUrl()
	if sshURL != "" {
		t.Fatalf("Want no url but got %s\n", sshURL)
	}
}
