package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"ghconfig/internal"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-github/v32/github"
)

func TestSync_SimpleSingleWorkflowUpdate(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "StarpTech"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "StarpTech in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "shikaka", "full_name": "StarpTech/shikaka", "owner": {"id":1, "Login": "StarpTech"}}]}`)
	})
	mux.HandleFunc("/repos/StarpTech/shikaka/contents/.github/workflows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[
		{
		  "type": "file",
		  "size": 20678,
		  "name": "ci.yaml",
		  "path": ".github/workflows/ci.yaml",
		  "sha": "sha"
		}]`)
	})
	mux.HandleFunc("/repos/StarpTech/shikaka/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/StarpTech/shikaka/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/StarpTech/shikaka/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/StarpTech/shikaka/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/StarpTech/shikaka/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/StarpTech/shikaka/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Update workflows by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/StarpTech/shikaka/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/StarpTech/shikaka/pull/20"}`)
	})
	mux.HandleFunc("/repos/StarpTech/shikaka/contents/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{
			"content":{
				"name":"p"
			},
			"commit":{
				"message":"m",
				"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
				"html_url": "https://github.com/StarpTech/shikaka/blob/master/.github/workflows/nodejs.yml"
			}
		}`)
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &internal.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "StarpTech",
		WorkflowRoot:    "workflows",
		RootDir:         "../test/fixture/simple-workflow",
	}

	getList := func(reposNames []string) []string {
		return []string{"StarpTech/shikaka"}
	}
	err := NewSyncCmd(cfg, RepositoryPrompt(getList))
	if err != nil {
		t.Fatalf("could not execute sync command, %v", err)
	}
}

func TestSync_SimpleSinglePatchUpdate(t *testing.T) {
	client, mux, serverURL, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "StarpTech"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "StarpTech in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "shikaka", "full_name": "StarpTech/shikaka", "owner": {"id":1, "Login": "StarpTech"}}]}`)
	})
	mux.HandleFunc("/repos/StarpTech/shikaka/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/StarpTech/shikaka/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/StarpTech/shikaka/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})
	mux.HandleFunc("/repos/StarpTech/shikaka/contents/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
		  "type": "file",
		  "name": "f",
		  "download_url": "`+serverURL+baseURLPath+`/download/.github/workflows/ci.yaml"
		}`)
	})
	mux.HandleFunc("/download/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, "name: foo")
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/StarpTech/shikaka/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/StarpTech/shikaka/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/StarpTech/shikaka/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Update workflows by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/StarpTech/shikaka/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/StarpTech/shikaka/pull/20"}`)
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &internal.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "StarpTech",
		WorkflowRoot:    "workflows",
		RootDir:         "../test/fixture/simple-patch",
	}

	getList := func(reposNames []string) []string {
		return []string{"StarpTech/shikaka"}
	}
	err := NewSyncCmd(cfg, RepositoryPrompt(getList))
	if err != nil {
		t.Fatalf("could not execute sync command, %v", err)
	}
}
