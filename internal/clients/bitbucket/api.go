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

var ErrNotFound = errors.New("not found")

const (
	PermissionRepoWrite = "REPO_WRITE"
	PermissionRepoRead  = "REPO_READ"
)

type AccessKey struct {
	Key        string
	Label      string
	ID         int
	Permission string
}

type Webhook struct {
	ID int `json:"id"`

	Name string `json:"name"`

	Configuration struct {
		Secret string `json:"secret"`
	} `json:"configuration"`

	Events []string `json:"events"`

	URL string `json:"url"`

	// active bool
}

// WebhookClientAPI is the API for creating/listing/deleting/getting webhooks
type WebhookClientAPI interface {
	CreateWebhook(ctx context.Context, repo Repo, webhook Webhook) (result Webhook, err error)
	DeleteWebhook(ctx context.Context, repo Repo, id int) (err error)
	GetWebhook(ctx context.Context, repo Repo, id int) (result Webhook, err error)
	//	ListAccessKeys(ctx context.Context, repo Repo) (result []AccessKey, err error)
	//	UpdateAccessKeyPermission(ctx context.Context, repo Repo, id int, permission string) (err error)
}
