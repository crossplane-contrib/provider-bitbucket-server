/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bitbucket

import (
	"context"
	"errors"
)

// Repo struct
type Repo struct {
	ProjectKey string
	Repo       string
}

// KeyClientAPI is the API for creating/listing/deleting/getting access keys
type KeyClientAPI interface {
	CreateAccessKey(ctx context.Context, repo Repo, key AccessKey) (result AccessKey, err error)
	DeleteAccessKey(ctx context.Context, repo Repo, id int) (err error)
	GetAccessKey(ctx context.Context, repo Repo, id int) (result AccessKey, err error)
	ListAccessKeys(ctx context.Context, repo Repo) (result []AccessKey, err error)
	UpdateAccessKeyPermission(ctx context.Context, repo Repo, id int, permission string) (err error)
}

// ErrNotFound returned when item is not found
var ErrNotFound = errors.New("not found")

const (
	// PermissionRepoWrite grants read write permissions to the repository
	PermissionRepoWrite = "REPO_WRITE"
	// PermissionRepoRead grants read only permissions to the repository
	PermissionRepoRead = "REPO_READ"
)

// AccessKey defines the api object for bitbucket server
type AccessKey struct {
	// Key is the public ssh key
	Key string
	// Label is the text description of the key
	Label string
	// ID is the number the access key is given by server
	ID int
	// Permission is either PermissionRepoRead or PermissionRepoWrite
	Permission string
}

// Webhook defines the api object for the bitbucket server objet webhook
type Webhook struct {
	// ID of the webhook in the server
	ID int `json:"id"`

	// Name of the webhook
	Name string `json:"name"`

	// Configuration contains webhook configurations
	Configuration struct {
		// Secret defines the authentication key that the bitbucket server HMAC signes the payload
		Secret string `json:"secret"`
	} `json:"configuration"`

	// Events defines for which events the webhook subscribes
	Events []string `json:"events"`

	// URL is the enpoint that bitbucket server should POST events to
	URL string `json:"url"`

	// active bool
}

// WebhookClientAPI is the API for creating/listing/deleting/getting webhooks
type WebhookClientAPI interface {
	CreateWebhook(ctx context.Context, repo Repo, webhook Webhook) (result Webhook, err error)
	DeleteWebhook(ctx context.Context, repo Repo, id int) (err error)
	GetWebhook(ctx context.Context, repo Repo, id int) (result Webhook, err error)
	UpdateWebhook(ctx context.Context, repo Repo, id int, webhook Webhook) (result Webhook, err error)
}
