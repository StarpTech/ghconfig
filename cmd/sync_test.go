package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"ghconfig/internal/config"
	"ghconfig/internal/dependabot"
	"net/http"
	"reflect"
	"testing"

	"github.com/apex/log"
	"github.com/apex/log/handlers/memory"
	"github.com/google/go-github/v32/github"
	"github.com/tj/assert"
	"gopkg.in/yaml.v2"
)

func TestSync_WorkflowNotExistOnRemote(t *testing.T) {
	client, mux, serverURL, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "o"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "o in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "r", "full_name": "o/r", "owner": {"id":1, "Login": "o"}}]}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[]`)
	})
	mux.HandleFunc("/repos/o/r/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/o/r/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/o/r/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Synchronize (.github) configurations by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/o/r/pull/20"}`)
	})

	mux.HandleFunc("/repos/o/r/contents/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, `{
				"type": "file",
				"name": "f",
				"path": ".github/workflows/ci.yaml",
				"download_url": "`+serverURL+baseURLPath+`/download/.github/workflows/ci.yaml"
			  }`)
		case "PUT":
			wf := readWorkflow(r.Body)
			assert.Equal(t, wf.Name, "Node CI")

			fmt.Fprint(w, `
			{
				"content":{
					"name":"CI"
				},
				"commit":{
					"message":"m",
					"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
					"html_url": "https://github.com/o/r/blob/master/.github/workflows/ci.yaml"
				}
			}`)
		default:
			t.Errorf("Request method: %v, want %v", r.Method, "PUT or GET")
		}
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &config.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "o in:name",
		RootDir:         "../test/fixture/simple-workflow",
	}

	h := memory.New()
	log.SetHandler(h)

	stub := StubRepositorySelection([]string{"o/r"})
	defer stub()

	err := NewSyncCmd(cfg)
	if err != nil {
		t.Fatalf("could not execute command, %v", err)
	}

	for _, entry := range h.Entries {
		if entry.Level >= log.ErrorLevel {
			t.Errorf("stderr should be empty, error: %v", entry)
		}
	}
}

func TestSync_WorkflowExistOnRemote(t *testing.T) {
	client, mux, serverURL, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "o"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "o in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "r", "full_name": "o/r", "owner": {"id":1, "Login": "o"}}]}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/repos/o/r/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/o/r/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/o/r/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Synchronize (.github) configurations by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/o/r/pull/20"}`)
	})

	mux.HandleFunc("/repos/o/r/contents/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, `{
				"type": "file",
				"name": "f",
				"path": ".github/workflows/ci.yaml",
				"download_url": "`+serverURL+baseURLPath+`/download/.github/workflows/ci.yaml"
			  }`)
		case "PUT":
			wf := readWorkflow(r.Body)
			assert.Equal(t, wf.Name, "Node CI")
			assert.Equal(t, wf.Env["foo"], "bar")
			assert.EqualValues(t, wf.Jobs["build"].Strategy.Matrix.NodeVersion, []string{"11.x", "12.x", "14.x"})
			assert.NotNil(t, wf.Jobs["build"])
			assert.Equal(t, len(wf.Jobs["build"].Steps), 4)

			fmt.Fprint(w, `
			{
				"content":{
					"name":"CI"
				},
				"commit":{
					"message":"m",
					"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
					"html_url": "https://github.com/o/r/blob/master/.github/workflows/ci.yaml"
				}
			}`)
		default:
			t.Errorf("Request method: %v, want %v", r.Method, "PUT or GET")
		}
	})
	mux.HandleFunc("/download/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"name": "patch",
			"env": { "foo": "bar" },
			"jobs": {
				"build": {
					"name": "build",
					"steps": [{ "name": "ferfr", "run": "npm install" }] 
				}
			}
		}`)
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &config.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "o in:name",
		RootDir:         "../test/fixture/simple-workflow",
	}

	h := memory.New()
	log.SetHandler(h)

	stub := StubRepositorySelection([]string{"o/r"})
	defer stub()

	err := NewSyncCmd(cfg)
	if err != nil {
		t.Fatalf("could not execute command, %v", err)
	}

	for _, entry := range h.Entries {
		if entry.Level >= log.ErrorLevel {
			t.Errorf("stderr should be empty, error: %v", entry)
		}
	}
}

func TestSync_WorkflowEmptyWorkflowFolderRemote(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "o"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "o in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "r", "full_name": "o/r", "owner": {"id":1, "Login": "o"}}]}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[]`)
	})
	mux.HandleFunc("/repos/o/r/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/o/r/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/o/r/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Synchronize (.github) configurations by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/o/r/pull/20"}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{
			"content":{
				"name":"p"
			},
			"commit":{
				"message":"m",
				"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
				"html_url": "https://github.com/o/r/blob/master/.github/workflows/nodejs.yml"
			}
		}`)
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &config.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "o in:name",
		RootDir:         "../test/fixture/simple-workflow",
	}

	h := memory.New()
	log.SetHandler(h)

	stub := StubRepositorySelection([]string{"o/r"})
	defer stub()

	err := NewSyncCmd(cfg)
	if err != nil {
		t.Fatalf("could not execute command, %v", err)
	}

	for _, entry := range h.Entries {
		if entry.Level >= log.ErrorLevel {
			t.Errorf("stderr should be empty, error: %v", entry)
		}
	}
}

func TestSync_Patch(t *testing.T) {
	client, mux, serverURL, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "o"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "o in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "r", "full_name": "o/r", "owner": {"id":1, "Login": "o"}}]}`)
	})
	mux.HandleFunc("/repos/o/r/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/o/r/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, `{
				"type": "file",
				"name": "f",
				"path": ".github/workflows/ci.yaml",
				"download_url": "`+serverURL+baseURLPath+`/download/.github/workflows/ci.yaml"
			  }`)
		case "PUT":
			wf := readWorkflow(r.Body)
			assert.Equal(t, wf.Name, "CI")
			fmt.Fprint(w, `
			{
				"content":{
					"name":"CI"
				},
				"commit":{
					"message":"m",
					"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
					"html_url": "https://github.com/o/r/blob/master/.github/workflows/ci.yml"
				}
			}`)
		default:
			t.Errorf("Request method: %v, want %v", r.Method, "PUT or GET")
		}
	})
	mux.HandleFunc("/download/.github/workflows/ci.yaml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, "name: foo")
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/o/r/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})

	input := &github.NewPullRequest{
		Title: github.String("Synchronize (.github) configurations by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/o/r/pull/20"}`)
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &config.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "o in:name",
		RootDir:         "../test/fixture/simple-patch",
	}

	h := memory.New()
	log.SetHandler(h)

	stub := StubRepositorySelection([]string{"o/r"})
	defer stub()

	err := NewSyncCmd(cfg)
	if err != nil {
		t.Fatalf("could not execute command, %v", err)
	}

	for _, entry := range h.Entries {
		if entry.Level >= log.ErrorLevel {
			t.Errorf("stderr should be empty, error: %v", entry)
		}
	}
}

func TestSync_DependabotNotExistOnRemote(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "o"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "o in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "r", "full_name": "o/r", "owner": {"id":1, "Login": "o"}}]}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[]`)
	})
	mux.HandleFunc("/repos/o/r/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/o/r/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/o/r/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Synchronize (.github) configurations by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/o/r/pull/20"}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/dependabot.yml", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		case "PUT":
			wf := readDependabot(r.Body)
			assert.Equal(t, wf.Version, "2")
			fmt.Fprint(w, `
			{
				"content":{
					"name":"CI"
				},
				"commit":{
					"message":"m",
					"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
					"html_url": "https://github.com/o/r/blob/master/.github/dependabot.yml"
				}
			}`)
		default:
			t.Errorf("Request method: %v, want %v", r.Method, "PUT or GET")
		}
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &config.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "o in:name",
		RootDir:         "../test/fixture/simple-dependabot",
	}

	h := memory.New()
	log.SetHandler(h)

	stub := StubRepositorySelection([]string{"o/r"})
	defer stub()

	err := NewSyncCmd(cfg)
	if err != nil {
		t.Fatalf("could not execute command, %v", err)
	}

	for _, entry := range h.Entries {
		if entry.Level >= log.ErrorLevel {
			t.Errorf("stderr should be empty, error: %v", entry)
		}
	}
}

func TestSync_DependabotExistOnRemote(t *testing.T) {
	client, mux, serverURL, teardown := setup()
	defer teardown()

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "Login": "o"}`)
	})
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{
			"q":        "o in:name",
			"page":     "1",
			"per_page": "120",
		})

		fmt.Fprint(w, `{"total_count": 1, "incomplete_results": false, "items": [{"id":1, "name": "r", "full_name": "o/r", "owner": {"id":1, "Login": "o"}}]}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/workflows", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[]`)
	})
	mux.HandleFunc("/repos/o/r/git/matching-refs/heads/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		  [
		    {
		      "ref": "refs/heads/master",
		      "url": "https://api.github.com/repos/o/r/git/refs/heads/master",
		      "object": {
		        "type": "commit",
		        "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		        "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		      }
		    }
		  ]`)
	})

	args := &createRefRequest{
		Ref: github.String("refs/heads/ghconfig/workflows/fixed_id"),
		SHA: github.String("aa218f56b14c9653891f9e74264a383fa43fefbd"),
	}
	mux.HandleFunc("/repos/o/r/git/refs", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRefRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, args) {
			t.Errorf("Request body = %+v, want %+v", v, args)
		}
		fmt.Fprint(w, `
		  {
		    "ref": "refs/heads/ghconfig/workflows/fixed_id",
		    "url": "https://api.github.com/repos/o/r/git/refs/heads/ghconfig/workflows/fixed_id",
		    "object": {
		      "type": "commit",
		      "sha": "aa218f56b14c9653891f9e74264a383fa43fefbd",
		      "url": "https://api.github.com/repos/o/r/git/commits/aa218f56b14c9653891f9e74264a383fa43fefbd"
		    }
		  }`)
	})
	input := &github.NewPullRequest{
		Title: github.String("Synchronize (.github) configurations by ghconfig"),
		Head:  github.String("ghconfig/workflows/fixed_id"),
		Base:  github.String("master"),
		Draft: github.Bool(true),
	}

	mux.HandleFunc("/repos/o/r/pulls", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.NewPullRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"number":1, "html_url": "https://github.com/o/r/pull/20"}`)
	})
	mux.HandleFunc("/repos/o/r/contents/.github/dependabot.yml", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, `{
				"type": "file",
				"name": "f",
				"path": ".github/dependabot.yml",
				"download_url": "`+serverURL+baseURLPath+`/download/.github/dependabot.yml"
			  }`)
		case "PUT":
			fmt.Fprint(w, `
			{
				"content":{
					"name":"CI"
				},
				"commit":{
					"message":"m",
					"sha":"f5f369044773ff9c6383c087466d12adb6fa0828",
					"html_url": "https://github.com/o/r/blob/master/.github/dependabot.yml"
				}
			}`)
		default:
			t.Errorf("Request method: %v, want %v", r.Method, "PUT or GET")
		}
	})
	mux.HandleFunc("/download/.github/dependabot.yml", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		bytes, _ := yaml.Marshal(&dependabot.GithubDependabot{
			Version: "1",
		})
		fmt.Fprint(w, string(bytes))
	})

	ctx := context.Background()
	sid := testIDGenerator{}

	cfg := &config.Config{
		GithubClient:    client,
		Context:         ctx,
		DryRun:          false,
		BaseBranch:      "master",
		Sid:             sid,
		CreatePR:        true,
		RepositoryQuery: "o in:name",
		RootDir:         "../test/fixture/simple-dependabot",
	}

	h := memory.New()
	log.SetHandler(h)

	stub := StubRepositorySelection([]string{"o/r"})
	defer stub()

	err := NewSyncCmd(cfg)
	if err != nil {
		t.Fatalf("could not execute command, %v", err)
	}

	for _, entry := range h.Entries {
		if entry.Level >= log.ErrorLevel {
			t.Errorf("stderr should be empty, error: %v", entry)
		}
	}
}
