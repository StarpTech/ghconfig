package dependabot

import "github.com/xeipuuv/gojsonschema"

var schemaLoader = gojsonschema.NewStringLoader(`{
	"$schema": "http://json-schema.org/draft-07/schema",
	"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions",
	"definitions": {
	  "architecture": {
		"type": "string",
		"enum": [
		  "ARM32",
		  "x64",
		  "x86"
		]
	  },
	  "branch": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#onpushpull_requestbranchestags",
		"$ref": "#/definitions/globs",
		"description": "When using the push and pull_request events, you can configure a workflow to run on specific branches or tags. If you only define only tags or only branches, the workflow won't run for events affecting the undefined Git ref.\nThe branches, branches-ignore, tags, and tags-ignore keywords accept glob patterns that use the * and ** wildcard characters to match more than one branch or tag name. For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet.\nThe patterns defined in branches and tags are evaluated against the Git ref's name. For example, defining the pattern mona/octocat in branches will match the refs/heads/mona/octocat Git ref. The pattern releases/** will match the refs/heads/releases/10 Git ref.\nYou can use two types of filters to prevent a workflow from running on pushes and pull requests to tags and branches:\n- branches or branches-ignore - You cannot use both the branches and branches-ignore filters for the same event in a workflow. Use the branches filter when you need to filter branches for positive matches and exclude branches. Use the branches-ignore filter when you only need to exclude branch names.\n- tags or tags-ignore - You cannot use both the tags and tags-ignore filters for the same event in a workflow. Use the tags filter when you need to filter tags for positive matches and exclude tags. Use the tags-ignore filter when you only need to exclude tag names.\nYou can exclude tags and branches using the ! character. The order that you define patterns matters.\n- A matching negative pattern (prefixed with !) after a positive match will exclude the Git ref.\n- A matching positive pattern after a negative match will include the Git ref again."
	  },
	  "configuration": {
		"oneOf": [
		  {
			"type": "string"
		  },
		  {
			"type": "number"
		  },
		  {
			"type": "boolean"
		  },
		  {
			"type": "object",
			"additionalProperties": {
			  "$ref": "#/definitions/configuration"
			}
		  },
		  {
			"type": "array",
			"items": {
			  "$ref": "#/definitions/configuration"
			},
			"additionalItems": false
		  }
		]
	  },
	  "container": {
		"oneOf": [
		  {
			"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainer",
			"description": "The Docker image to use as the container to run the action. The value can be the Docker Hub image name or a public docker registry name.",
			"type": "string"
		  },
		  {
			"type": "object",
			"properties": {
			  "image": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainerimage",
				"description": "The Docker image to use as the container to run the action. The value can be the Docker Hub image name or a public docker registry name.",
				"type": "string"
			  },
			  "env": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainerenv",
				"$ref": "#/definitions/env",
				"description": "Sets an array of environment variables in the container."
			  },
			  "ports": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainerports",
				"description": "Sets an array of ports to expose on the container.",
				"type": "array",
				"items": {
				  "oneOf": [
					{
					  "type": "number"
					},
					{
					  "type": "string"
					}
				  ]
				},
				"minItems": 1,
				"additionalItems": false
			  },
			  "volumes": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainervolumes",
				"description": "Sets an array of volumes for the container to use. You can use volumes to share data between services or other steps in a job. You can specify named Docker volumes, anonymous Docker volumes, or bind mounts on the host.\nTo specify a volume, you specify the source and destination path: <source>:<destinationPath>\nThe <source> is a volume name or an absolute path on the host machine, and <destinationPath> is an absolute path in the container.",
				"type": "array",
				"items": {
				  "type": "string",
				  "pattern": "^[^:]+:[^:]+$"
				},
				"minItems": 1,
				"additionalItems": false
			  },
			  "options": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontaineroptions",
				"description": "Additional Docker container resource options. For a list of options, see https://docs.docker.com/engine/reference/commandline/create/#options.",
				"type": "string"
			  }
			},
			"required": [
			  "image"
			],
			"additionalProperties": false
		  }
		]
	  },
	  "defaults": {
		"type": "object",
		"properties": {
		  "run": {
			"type": "object",
			"properties": {
			  "shell": {
				"$ref": "#/definitions/shell"
			  },
			  "working-directory": {
				"$ref": "#/definitions/working-directory"
			  }
			},
			"minProperties": 1,
			"additionalProperties": false
		  }
		},
		"minProperties": 1,
		"additionalProperties": false
	  },
	  "env": {
		"type": "object",
		"additionalProperties": {
		  "oneOf": [
			{
			  "type": "string"
			},
			{
			  "type": "number"
			},
			{
			  "type": "boolean"
			}
		  ]
		},
		"minProperties": 1
	  },
	  "event": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows",
		"type": "string",
		"enum": [
		  "check_run",
		  "check_suite",
		  "create",
		  "delete",
		  "deployment",
		  "deployment_status",
		  "fork",
		  "gollum",
		  "issue_comment",
		  "issues",
		  "label",
		  "member",
		  "milestone",
		  "page_build",
		  "project",
		  "project_card",
		  "project_column",
		  "public",
		  "pull_request",
		  "pull_request_review",
		  "pull_request_review_comment",
		  "push",
		  "registry_package",
		  "release",
		  "status",
		  "watch",
		  "workflow_dispatch",
		  "workflow_run",
		  "repository_dispatch"
		]
	  },
	  "eventObject": {
		"oneOf": [
		  {
			"type": "object"
		  },
		  {
			"type": "null"
		  }
		],
		"additionalProperties": true
	  },
	  "expressionSyntax": {
		"type": "string",
		"pattern": "^\\${{.*}}$"
	  },
	  "globs": {
		"type": "array",
		"items": {
		  "type": "string",
		  "minLength": 1
		},
		"minItems": 1,
		"additionalItems": false
	  },
	  "machine": {
		"type": "string",
		"enum": [
		  "linux",
		  "macos",
		  "windows"
		]
	  },
	  "name": {
		"type": "string",
		"pattern": "^[_a-zA-Z][a-zA-Z0-9_-]*$"
	  },
	  "path": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#onpushpull_requestpaths",
		"$ref": "#/definitions/globs",
		"description": "When using the push and pull_request events, you can configure a workflow to run when at least one file does not match paths-ignore or at least one modified file matches the configured paths. Path filters are not evaluated for pushes to tags.\nThe paths-ignore and paths keywords accept glob patterns that use the * and ** wildcard characters to match more than one path name. For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#filter-pattern-cheat-sheet.\nYou can exclude paths using two types of filters. You cannot use both of these filters for the same event in a workflow.\n- paths-ignore - Use the paths-ignore filter when you only need to exclude path names.\n- paths - Use the paths filter when you need to filter paths for positive matches and exclude paths."
	  },
	  "ref": {
		"properties": {
		  "branches": {
			"$ref": "#/definitions/branch"
		  },
		  "branches-ignore": {
			"$ref": "#/definitions/branch"
		  },
		  "tags": {
			"$ref": "#/definitions/branch"
		  },
		  "tags-ignore": {
			"$ref": "#/definitions/branch"
		  },
		  "paths": {
			"$ref": "#/definitions/path"
		  },
		  "paths-ignore": {
			"$ref": "#/definitions/path"
		  }
		},
		"oneOf": [
		  {
			"type": "object",
			"allOf": [
			  {
				"not": {
				  "required": [
					"branches",
					"branches-ignore"
				  ]
				}
			  },
			  {
				"not": {
				  "required": [
					"tags",
					"tags-ignore"
				  ]
				}
			  },
			  {
				"not": {
				  "required": [
					"paths",
					"paths-ignore"
				  ]
				}
			  }
			]
		  },
		  {
			"type": "null"
		  }
		]
	  },
	  "shell": {
		"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#using-a-specific-shell",
		"description": "You can override the default shell settings in the runner's operating system using the shell keyword. You can use built-in shell keywords, or you can define a custom set of shell options.",
		"type": "string",
		"anyOf": [
		  {
			"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#custom-shell"
		  },
		  {
			"enum": [
			  "bash",
			  "pwsh",
			  "python",
			  "sh",
			  "cmd",
			  "powershell"
			]
		  }
		]
	  },
	  "types": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#onevent_nametypes",
		"description": "Selects the types of activity that will trigger a workflow run. Most GitHub events are triggered by more than one type of activity. For example, the event for the release resource is triggered when a release is published, unpublished, created, edited, deleted, or prereleased. The types keyword enables you to narrow down activity that causes the workflow to run. When only one activity type triggers a webhook event, the types keyword is unnecessary.\nYou can use an array of event types. For more information about each event and their activity types, see https://help.github.com/en/articles/events-that-trigger-workflows#webhook-events.",
		"type": "array",
		"minItems": 1,
		"additionalItems": false
	  },
	  "working-directory": {
		"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsrun",
		"description": "Using the working-directory keyword, you can specify the working directory of where to run the command.",
		"type": "string"
	  }
	},
	"properties": {
	  "name": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#name",
		"description": "The name of your workflow. GitHub displays the names of your workflows on your repository's actions page. If you omit this field, GitHub sets the name to the workflow's filename.",
		"type": "string"
	  },
	  "on": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#on",
		"description": "The name of the GitHub event that triggers the workflow. You can provide a single event string, array of events, array of event types, or an event configuration map that schedules a workflow or restricts the execution of a workflow to specific files, tags, or branch changes. For a list of available events, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows.",
		"oneOf": [
		  {
			"$ref": "#/definitions/event"
		  },
		  {
			"type": "array",
			"items": {
			  "$ref": "#/definitions/event"
			},
			"minItems": 1,
			"additionalItems": false
		  },
		  {
			"type": "object",
			"properties": {
			  "check_run": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#check-run-event-check_run",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the check_run event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/checks/runs.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"rerequested",
						"completed",
						"requested_action"
					  ]
					},
					"default": [
					  "created",
					  "rerequested",
					  "completed",
					  "requested_action"
					]
				  }
				}
			  },
			  "check_suite": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#check-suite-event-check_suite",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the check_suite event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/checks/suites/.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"completed",
						"requested",
						"rerequested"
					  ]
					},
					"default": [
					  "completed",
					  "requested",
					  "rerequested"
					]
				  }
				}
			  },
			  "create": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#create-event-create",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime someone creates a branch or tag, which triggers the create event. For information about the REST API, see https://developer.github.com/v3/git/refs/#create-a-reference."
			  },
			  "delete": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#delete-event-delete",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime someone deletes a branch or tag, which triggers the delete event. For information about the REST API, see https://developer.github.com/v3/git/refs/#delete-a-reference."
			  },
			  "deployment": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#deployment-event-deployment",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime someone creates a deployment, which triggers the deployment event. Deployments created with a commit SHA may not have a Git ref. For information about the REST API, see https://developer.github.com/v3/repos/deployments/."
			  },
			  "deployment_status": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#deployment-status-event-deployment_status",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime a third party provides a deployment status, which triggers the deployment_status event. Deployments created with a commit SHA may not have a Git ref. For information about the REST API, see https://developer.github.com/v3/repos/deployments/#create-a-deployment-status."
			  },
			  "fork": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#fork-event-fork",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime when someone forks a repository, which triggers the fork event. For information about the REST API, see https://developer.github.com/v3/repos/forks/#create-a-fork."
			  },
			  "gollum": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#gollum-event-gollum",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow when someone creates or updates a Wiki page, which triggers the gollum event."
			  },
			  "issue_comment": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#issue-comment-event-issue_comment",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the issue_comment event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/issues/comments/.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "issues": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#issues-event-issues",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the issues event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/issues.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"opened",
						"edited",
						"deleted",
						"transferred",
						"pinned",
						"unpinned",
						"closed",
						"reopened",
						"assigned",
						"unassigned",
						"labeled",
						"unlabeled",
						"locked",
						"unlocked",
						"milestoned",
						"demilestoned"
					  ]
					},
					"default": [
					  "opened",
					  "edited",
					  "deleted",
					  "transferred",
					  "pinned",
					  "unpinned",
					  "closed",
					  "reopened",
					  "assigned",
					  "unassigned",
					  "labeled",
					  "unlabeled",
					  "locked",
					  "unlocked",
					  "milestoned",
					  "demilestoned"
					]
				  }
				}
			  },
			  "label": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#label-event-label",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the label event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/issues/labels/.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "member": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#member-event-member",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the member event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/repos/collaborators/.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"added",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "added",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "milestone": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#milestone-event-milestone",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the milestone event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/issues/milestones/.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"closed",
						"opened",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "closed",
					  "opened",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "page_build": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#page-build-event-page_build",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime someone pushes to a GitHub Pages-enabled branch, which triggers the page_build event. For information about the REST API, see https://developer.github.com/v3/repos/pages/."
			  },
			  "project": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#project-event-project",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the project event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/projects/.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"updated",
						"closed",
						"reopened",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "updated",
					  "closed",
					  "reopened",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "project_card": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#project-card-event-project_card",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the project_card event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/projects/cards.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"moved",
						"converted",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "moved",
					  "converted",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "project_column": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#project-column-event-project_column",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the project_column event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/projects/columns.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"updated",
						"moved",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "updated",
					  "moved",
					  "deleted"
					]
				  }
				}
			  },
			  "public": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#public-event-public",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime someone makes a private repository public, which triggers the public event. For information about the REST API, see https://developer.github.com/v3/repos/#edit."
			  },
			  "pull_request": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#pull-request-event-pull_request",
				"$ref": "#/definitions/ref",
				"description": "Runs your workflow anytime the pull_request event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/pulls.\nNote: Workflows do not run on private base repositories when you open a pull request from a forked repository.\nWhen you create a pull request from a forked repository to the base repository, GitHub sends the pull_request event to the base repository and no pull request events occur on the forked repository.\nWorkflows don't run on forked repositories by default. You must enable GitHub Actions in the Actions tab of the forked repository.\nThe permissions for the GITHUB_TOKEN in forked repositories is read-only. For more information about the GITHUB_TOKEN, see https://help.github.com/en/articles/virtual-environments-for-github-actions.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"assigned",
						"unassigned",
						"labeled",
						"unlabeled",
						"opened",
						"edited",
						"closed",
						"reopened",
						"synchronize",
						"ready_for_review",
						"locked",
						"unlocked",
						"review_requested",
						"review_request_removed"
					  ]
					},
					"default": [
					  "opened",
					  "synchronize",
					  "reopened"
					]
				  }
				},
				"patternProperties": {
				  "^(branche|tag|path)s(-ignore)?$": {}
				},
				"additionalProperties": false
			  },
			  "pull_request_review": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#pull-request-review-event-pull_request_review",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the pull_request_review event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/pulls/reviews.\nNote: Workflows do not run on private base repositories when you open a pull request from a forked repository.\nWhen you create a pull request from a forked repository to the base repository, GitHub sends the pull_request event to the base repository and no pull request events occur on the forked repository.\nWorkflows don't run on forked repositories by default. You must enable GitHub Actions in the Actions tab of the forked repository.\nThe permissions for the GITHUB_TOKEN in forked repositories is read-only. For more information about the GITHUB_TOKEN, see https://help.github.com/en/articles/virtual-environments-for-github-actions.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"submitted",
						"edited",
						"dismissed"
					  ]
					},
					"default": [
					  "submitted",
					  "edited",
					  "dismissed"
					]
				  }
				}
			  },
			  "pull_request_review_comment": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#pull-request-review-comment-event-pull_request_review_comment",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime a comment on a pull request's unified diff is modified, which triggers the pull_request_review_comment event. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/pulls/comments.\nNote: Workflows do not run on private base repositories when you open a pull request from a forked repository.\nWhen you create a pull request from a forked repository to the base repository, GitHub sends the pull_request event to the base repository and no pull request events occur on the forked repository.\nWorkflows don't run on forked repositories by default. You must enable GitHub Actions in the Actions tab of the forked repository.\nThe permissions for the GITHUB_TOKEN in forked repositories is read-only. For more information about the GITHUB_TOKEN, see https://help.github.com/en/articles/virtual-environments-for-github-actions.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"created",
						"edited",
						"deleted"
					  ]
					},
					"default": [
					  "created",
					  "edited",
					  "deleted"
					]
				  }
				}
			  },
			  "push": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#push-event-push",
				"$ref": "#/definitions/ref",
				"description": "Runs your workflow when someone pushes to a repository branch, which triggers the push event.\nNote: The webhook payload available to GitHub Actions does not include the added, removed, and modified attributes in the commit object. You can retrieve the full commit object using the REST API. For more information, see https://developer.github.com/v3/repos/commits/#get-a-single-commit.",
				"patternProperties": {
				  "^(branche|tag|path)s(-ignore)?$": {}
				},
				"additionalProperties": false
			  },
			  "registry_package": {
				"$comment": "https://help.github.com/en/actions/reference/events-that-trigger-workflows#registry-package-event-registry_package",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime a package is published or updated. For more information, see https://help.github.com/en/github/managing-packages-with-github-packages.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"published",
						"updated"
					  ]
					},
					"default": [
					  "published",
					  "updated"
					]
				  }
				}
			  },
			  "release": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#release-event-release",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the release event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/repos/releases/ in the GitHub Developer documentation.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"published",
						"unpublished",
						"created",
						"edited",
						"deleted",
						"prereleased",
						"released"
					  ]
					},
					"default": [
					  "published",
					  "unpublished",
					  "created",
					  "edited",
					  "deleted",
					  "prereleased",
					  "released"
					]
				  }
				}
			  },
			  "status": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#status-event-status",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the status of a Git commit changes, which triggers the status event. For information about the REST API, see https://developer.github.com/v3/repos/statuses/."
			  },
			  "watch": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#watch-event-watch",
				"$ref": "#/definitions/eventObject",
				"description": "Runs your workflow anytime the watch event occurs. More than one activity type triggers this event. For information about the REST API, see https://developer.github.com/v3/activity/starring/."
			  },
			  "workflow_dispatch": {
				"$comment": "https://github.blog/changelog/2020-07-06-github-actions-manual-triggers-with-workflow_dispatch/",
				"description": "You can now create workflows that are manually triggered with the new workflow_dispatch event. You will then see a 'Run workflow' button on the Actions tab, enabling you to easily trigger a run.",
				"properties": {
				  "inputs": {
					"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/metadata-syntax-for-github-actions#inputs",
					"description": "Input parameters allow you to specify data that the action expects to use during runtime. GitHub stores input parameters as environment variables. Input ids with uppercase letters are converted to lowercase during runtime. We recommended using lowercase input ids.",
					"type": "object",
					"patternProperties": {
					  "^[_a-zA-Z][a-zA-Z0-9_-]*$": {
						"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/metadata-syntax-for-github-actions#inputsinput_id",
						"description": "A string identifier to associate with the input. The value of <input_id> is a map of the input's metadata. The <input_id> must be a unique identifier within the inputs object. The <input_id> must start with a letter or _ and contain only alphanumeric characters, -, or _.",
						"type": "object",
						"properties": {
						  "description": {
							"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/metadata-syntax-for-github-actions#inputsinput_iddescription",
							"description": "A string description of the input parameter.",
							"type": "string"
						  },
						  "deprecationMessage": {
							"description": "A string shown to users using the deprecated input.",
							"type": "string"
						  },
						  "required": {
							"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/metadata-syntax-for-github-actions#inputsinput_idrequired",
							"description": "A boolean to indicate whether the action requires the input parameter. Set to true when the parameter is required.",
							"type": "boolean"
						  },
						  "default": {
							"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/metadata-syntax-for-github-actions#inputsinput_iddefault",
							"description": "A string representing the default value. The default value is used when an input parameter isn't specified in a workflow file.",
							"type": "string"
						  }
						},
						"required": [
						  "description",
						  "required"
						],
						"additionalProperties": false
					  }
					},
					"additionalProperties": false
				  }
				}
			  },
			  "workflow_run": {
				"$comment": "https://docs.github.com/en/actions/reference/events-that-trigger-workflows#workflow_run",
				"$ref": "#/definitions/eventObject",
				"description": "This event occurs when a workflow run is requested or completed, and allows you to execute a workflow based on the finished result of another workflow. For example, if your pull_request workflow generates build artifacts, you can create a new workflow that uses workflow_run to analyze the results and add a comment to the original pull request.",
				"properties": {
				  "types": {
					"$ref": "#/definitions/types",
					"items": {
					  "type": "string",
					  "enum": [
						"requested",
						"completed"
					  ]
					},
					"default": [
					  "requested",
					  "completed"
					]
				  },
				  "workflows": {
					"type": "array",
					"items": {
					  "type": "string"
					},
					"minItems": 1,
					"additionalItems": false
				  }
				},
				"patternProperties": {
				  "^branches(-ignore)?$": {}
				}
			  },
			  "repository_dispatch": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#external-events-repository_dispatch",
				"$ref": "#/definitions/eventObject",
				"description": "You can use the GitHub API to trigger a webhook event called repository_dispatch when you want to trigger a workflow for activity that happens outside of GitHub. For more information, see https://developer.github.com/v3/repos/#create-a-repository-dispatch-event.\nTo trigger the custom repository_dispatch webhook event, you must send a POST request to a GitHub API endpoint and provide an event_type name to describe the activity type. To trigger a workflow run, you must also configure your workflow to use the repository_dispatch event."
			  },
			  "schedule": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows#scheduled-events-schedule",
				"description": "You can schedule a workflow to run at specific UTC times using POSIX cron syntax (https://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html#tag_20_25_07). Scheduled workflows run on the latest commit on the default or base branch. The shortest interval you can run scheduled workflows is once every 5 minutes.\nNote: GitHub Actions does not support the non-standard syntax @yearly, @monthly, @weekly, @daily, @hourly, and @reboot.\nYou can use crontab guru (https://crontab.guru/). to help generate your cron syntax and confirm what time it will run. To help you get started, there is also a list of crontab guru examples (https://crontab.guru/examples.html).",
				"type": "array",
				"items": {
				  "properties": {
					"cron": {
					  "$comment": "https://stackoverflow.com/a/57639657/4044345",
					  "type": "string",
					  "pattern": "^(((\\d+,)+\\d+|((\\d+|\\*)\\/\\d+)|(\\d+-\\d+)|\\d+|\\*) ?){5,7}$"
					}
				  },
				  "additionalProperties": false
				},
				"minItems": 1,
				"additionalItems": false
			  }
			},
			"additionalProperties": false
		  }
		]
	  },
	  "env": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#env",
		"$ref": "#/definitions/env",
		"description": "A map of environment variables that are available to all jobs and steps in the workflow."
	  },
	  "defaults": {
		"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#defaults",
		"$ref": "#/definitions/defaults",
		"description": "A map of default settings that will apply to all jobs in the workflow."
	  },
	  "jobs": {
		"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobs",
		"description": "A workflow run is made up of one or more jobs. Jobs run in parallel by default. To run jobs sequentially, you can define dependencies on other jobs using the jobs.<job_id>.needs keyword.\nEach job runs in a fresh instance of the virtual environment specified by runs-on.\nYou can run an unlimited number of jobs as long as you are within the workflow usage limits. For more information, see https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#usage-limits.",
		"type": "object",
		"patternProperties": {
		  "^[_a-zA-Z][a-zA-Z0-9_-]*$": {
			"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_id",
			"description": "Each job must have an id to associate with the job. The key job_id is a string and its value is a map of the job's configuration data. You must replace <job_id> with a string that is unique to the jobs object. The <job_id> must start with a letter or _ and contain only alphanumeric characters, -, or _.",
			"type": "object",
			"properties": {
			  "name": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idname",
				"description": "The name of the job displayed on GitHub.",
				"type": "string"
			  },
			  "needs": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idneeds",
				"description": "Identifies any jobs that must complete successfully before this job will run. It can be a string or array of strings. If a job fails, all jobs that need it are skipped unless the jobs use a conditional statement that causes the job to continue.",
				"oneOf": [
				  {
					"type": "array",
					"items": {
					  "$ref": "#/definitions/name"
					},
					"minItems": 1,
					"additionalItems": false
				  },
				  {
					"$ref": "#/definitions/name"
				  }
				]
			  },
			  "runs-on": {
				"$comment": "https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idruns-on",
				"description": "The type of machine to run the job on. The machine can be either a GitHub-hosted runner, or a self-hosted runner.",
				"oneOf": [
				  {
					"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#github-hosted-runners",
					"type": "string",
					"enum": [
					  "${{ matrix.os }}",
					  "macos-latest",
					  "macos-10.15",
					  "self-hosted",
					  "ubuntu-16.04",
					  "ubuntu-18.04",
					  "ubuntu-20.04",
					  "ubuntu-latest",
					  "windows-latest",
					  "windows-2019"
					]
				  },
				  {
					"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#self-hosted-runners",
					"type": "array",
					"anyOf": [
					  {
						"items": [
						  {
							"const": "self-hosted"
						  }
						]
					  },
					  {
						"items": [
						  {
							"const": "self-hosted"
						  },
						  {
							"$ref": "#/definitions/machine"
						  }
						]
					  },
					  {
						"items": [
						  {
							"const": "self-hosted"
						  },
						  {
							"$ref": "#/definitions/architecture"
						  }
						]
					  },
					  {
						"items": [
						  {
							"const": "self-hosted"
						  },
						  {
							"$ref": "#/definitions/machine"
						  },
						  {
							"$ref": "#/definitions/architecture"
						  }
						]
					  },
					  {
						"items": [
						  {
							"const": "self-hosted"
						  },
						  {
							"$ref": "#/definitions/architecture"
						  },
						  {
							"$ref": "#/definitions/machine"
						  }
						]
					  }
					],
					"additionalItems": {
					  "type": "string"
					}
				  }
				]
			  },
			  "outputs": {
				"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjobs_idoutputs",
				"description": "A map of outputs for a job. Job outputs are available to all downstream jobs that depend on this job.",
				"type": "object",
				"additionalProperties": {
				  "type": "string"
				},
				"minProperties": 1
			  },
			  "env": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idenv",
				"$ref": "#/definitions/env",
				"description": "A map of environment variables that are available to all steps in the job."
			  },
			  "defaults": {
				"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_iddefaults",
				"$ref": "#/definitions/defaults",
				"description": "A map of default settings that will apply to all steps in the job."
			  },
			  "if": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idif",
				"description": "You can use the if conditional to prevent a job from running unless a condition is met. You can use any supported context and expression to create a conditional.\nExpressions in an if conditional do not require the ${{ }} syntax. For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.",
				"type": "string"
			  },
			  "steps": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idsteps",
				"description": "A job contains a sequence of tasks called steps. Steps can run commands, run setup tasks, or run an action in your repository, a public repository, or an action published in a Docker registry. Not all steps run actions, but all actions run as a step. Each step runs in its own process in the virtual environment and has access to the workspace and filesystem. Because steps run in their own process, changes to environment variables are not preserved between steps. GitHub provides built-in steps to set up and complete a job.\n",
				"type": "array",
				"items": {
				  "type": "object",
				  "properties": {
					"id": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsid",
					  "description": "A unique identifier for the step. You can use the id to reference the step in contexts. For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.",
					  "type": "string"
					},
					"if": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsif",
					  "description": "You can use the if conditional to prevent a step from running unless a condition is met. You can use any supported context and expression to create a conditional.\nExpressions in an if conditional do not require the ${{ }} syntax. For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.",
					  "type": "string"
					},
					"name": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsname",
					  "description": "A name for your step to display on GitHub.",
					  "type": "string"
					},
					"uses": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsuses",
					  "description": "Selects an action to run as part of a step in your job. An action is a reusable unit of code. You can use an action defined in the same repository as the workflow, a public repository, or in a published Docker container image (https://hub.docker.com/).\nWe strongly recommend that you include the version of the action you are using by specifying a Git ref, SHA, or Docker tag number. If you don't specify a version, it could break your workflows or cause unexpected behavior when the action owner publishes an update.\n- Using the commit SHA of a released action version is the safest for stability and security.\n- Using the specific major action version allows you to receive critical fixes and security patches while still maintaining compatibility. It also assures that your workflow should still work.\n- Using the master branch of an action may be convenient, but if someone releases a new major version with a breaking change, your workflow could break.\nSome actions require inputs that you must set using the with keyword. Review the action's README file to determine the inputs required.\nActions are either JavaScript files or Docker containers. If the action you're using is a Docker container you must run the job in a Linux virtual environment. For more details, see https://help.github.com/en/articles/virtual-environments-for-github-actions.",
					  "type": "string"
					},
					"run": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsrun",
					  "description": "Runs command-line programs using the operating system's shell. If you do not provide a name, the step name will default to the text specified in the run command.\nCommands run using non-login shells by default. You can choose a different shell and customize the shell used to run commands. For more information, see https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell.\nEach run keyword represents a new process and shell in the virtual environment. When you provide multi-line commands, each line runs in the same shell.",
					  "type": "string"
					},
					"working-directory": {
					  "$ref": "#/definitions/working-directory"
					},
					"shell": {
					  "$ref": "#/definitions/shell"
					},
					"with": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswith",
					  "$ref": "#/definitions/env",
					  "description": "A map of the input parameters defined by the action. Each input parameter is a key/value pair. Input parameters are set as environment variables. The variable is prefixed with INPUT_ and converted to upper case.",
					  "properties": {
						"args": {
						  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswithargs",
						  "type": "string"
						},
						"entrypoint": {
						  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint",
						  "type": "string"
						}
					  }
					},
					"env": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepsenv",
					  "$ref": "#/definitions/env",
					  "description": "Sets environment variables for steps to use in the virtual environment. You can also set environment variables for the entire workflow or a job."
					},
					"continue-on-error": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepscontinue-on-error",
					  "description": "Prevents a job from failing when a step fails. Set to true to allow a job to pass when this step fails.",
					  "oneOf": [
						{
						  "type": "boolean"
						},
						{
						  "$ref": "#/definitions/expressionSyntax"
						}
					  ],
					  "default": false
					},
					"timeout-minutes": {
					  "$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes",
					  "description": "The maximum number of minutes to run the step before killing the process.",
					  "type": "number"
					}
				  },
				  "dependencies": {
					"working-directory": [
					  "run"
					],
					"shell": [
					  "run"
					]
				  },
				  "additionalProperties": false
				},
				"minItems": 1,
				"additionalItems": false
			  },
			  "timeout-minutes": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes",
				"description": "The maximum number of minutes to let a workflow run before GitHub automatically cancels it. Default: 360",
				"type": "number",
				"default": 360
			  },
			  "strategy": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategy",
				"description": "A strategy creates a build matrix for your jobs. You can define different variations of an environment to run each job in.",
				"type": "object",
				"properties": {
				  "matrix": {
					"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix",
					"description": "A build matrix is a set of different configurations of the virtual environment. For example you might run a job against more than one supported version of a language, operating system, or tool. Each configuration is a copy of the job that runs and reports a status.\nYou can specify a matrix by supplying an array for the configuration options. For example, if the GitHub virtual environment supports Node.js versions 6, 8, and 10 you could specify an array of those versions in the matrix.\nWhen you define a matrix of operating systems, you must set the required runs-on keyword to the operating system of the current job, rather than hard-coding the operating system name. To access the operating system name, you can use the matrix.os context parameter to set runs-on. For more information, see https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.",
					"oneOf": [
					  {
						"type": "object"
					  },
					  {
						"$ref": "#/definitions/expressionSyntax"
					  }
					],
					"patternProperties": {
					  "^(in|ex)clude$": {
						"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#example-including-configurations-in-a-matrix-build",
						"type": "array",
						"items": {
						  "type": "object",
						  "additionalProperties": {
							"$ref": "#/definitions/configuration"
						  }
						},
						"minItems": 1,
						"additionalItems": false
					  }
					},
					"additionalProperties": {
					  "type": "array",
					  "items": {
						"$ref": "#/definitions/configuration"
					  },
					  "minItems": 1,
					  "additionalItems": false
					},
					"minProperties": 1
				  },
				  "fail-fast": {
					"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategyfail-fast",
					"description": "When set to true, GitHub cancels all in-progress jobs if any matrix job fails. Default: true",
					"type": "boolean",
					"default": true
				  },
				  "max-parallel": {
					"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idstrategymax-parallel",
					"description": "The maximum number of jobs that can run simultaneously when using a matrix job strategy. By default, GitHub will maximize the number of jobs run in parallel depending on the available runners on GitHub-hosted virtual machines.",
					"type": "number"
				  }
				},
				"required": [
				  "matrix"
				],
				"additionalProperties": false
			  },
			  "continue-on-error": {
				"$comment": "https://help.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontinue-on-error",
				"description": "Prevents a workflow run from failing when a job fails. Set to true to allow a workflow run to pass when this job fails.",
				"oneOf": [
				  {
					"type": "boolean"
				  },
				  {
					"$ref": "#/definitions/expressionSyntax"
				  }
				]
			  },
			  "container": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idcontainer",
				"description": "A container to run any steps in a job that don't already specify a container. If you have steps that use both script and container actions, the container actions will run as sibling containers on the same network with the same volume mounts.\nIf you do not set a container, all steps will run directly on the host specified by runs-on unless a step refers to an action configured to run in a container.",
				"type": "object",
				"additionalProperties": {
				  "oneOf": [
					{
					  "type": "string"
					},
					{
					  "$ref": "#/definitions/container"
					}
				  ]
				}
			  },
			  "services": {
				"$comment": "https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#jobsjob_idservices",
				"description": "Additional containers to host services for a job in a workflow. These are useful for creating databases or cache services like redis. The runner on the virtual machine will automatically create a network and manage the life cycle of the service containers.\nWhen you use a service container for a job or your step uses container actions, you don't need to set port information to access the service. Docker automatically exposes all ports between containers on the same network.\nWhen both the job and the action run in a container, you can directly reference the container by its hostname. The hostname is automatically mapped to the service name.\nWhen a step does not use a container action, you must access the service using localhost and bind the ports.",
				"type": "object",
				"additionalProperties": {
				  "$ref": "#/definitions/container"
				}
			  }
			},
			"required": [
			  "runs-on"
			],
			"additionalProperties": false
		  }
		},
		"minProperties": 1,
		"additionalProperties": false
	  }
	},
	"required": [
	  "on",
	  "jobs"
	],
	"additionalProperties": false
  }`)

func ValidateSchema(json string) *gojsonschema.Result {
	documentLoader := gojsonschema.NewStringLoader(json)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}
	return result
}
