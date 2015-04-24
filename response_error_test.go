package stash

import (
	"errors"
	"net/http"
	"testing"
)

func TestRepositoryNotExists(t *testing.T) {
	if IsRepositoryExists(nil) {
		t.Fatalf("nil an errorResponse type")
	}

	if IsRepositoryExists(errors.New("foo")) {
		t.Fatalf("Not an errorResponse type")
	}

	if !IsRepositoryExists(errorResponse{StatusCode: http.StatusConflict}) {
		t.Fatalf("Want errorResponse.409")
	}

	if IsRepositoryExists(errorResponse{StatusCode: http.StatusNotFound}) {
		t.Fatalf("Want errorResponse.409")
	}
}

func TestRepositoryNotFound(t *testing.T) {
	if IsRepositoryNotFound(nil) {
		t.Fatalf("nil not an errorResponse type")
	}

	if IsRepositoryExists(errors.New("foo")) {
		t.Fatalf("Not an errorResponse type")
	}

	if !IsRepositoryNotFound(errorResponse{StatusCode: http.StatusNotFound}) {
		t.Fatalf("Want errorResponse.404")
	}

	if IsRepositoryNotFound(errorResponse{StatusCode: http.StatusConflict}) {
		t.Fatalf("Want errorResponse.404")
	}
}
