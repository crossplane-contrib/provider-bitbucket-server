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

package fake

import (
	"context"

	"github.com/crossplane/provider-bitbucket-server/internal/clients/bitbucket"
)

var _ bitbucket.WebhookClientAPI = &MockWebhookClient{}

// MockWebhookClient is a fake implementation of WebhookClientAPI
type MockWebhookClient struct {
	bitbucket.WebhookClientAPI

	MockCreateWebhook func(ctx context.Context, repo bitbucket.Repo, hook bitbucket.Webhook) (result bitbucket.Webhook, err error)
	MockDeleteWebhook func(ctx context.Context, repo bitbucket.Repo, id int) (err error)
	MockGetWebhook    func(ctx context.Context, repo bitbucket.Repo, id int) (result bitbucket.Webhook, err error)
	MockUpdateWebhook func(ctx context.Context, repo bitbucket.Repo, id int, hook bitbucket.Webhook) (result bitbucket.Webhook, err error)
}

// CreateWebhook calls the mock
func (c *MockWebhookClient) CreateWebhook(ctx context.Context, repo bitbucket.Repo, hook bitbucket.Webhook) (result bitbucket.Webhook, err error) {
	return c.MockCreateWebhook(ctx, repo, hook)
}

// DeleteWebhook calls the mock
func (c *MockWebhookClient) DeleteWebhook(ctx context.Context, repo bitbucket.Repo, id int) (err error) {
	return c.MockDeleteWebhook(ctx, repo, id)
}

// GetWebhook calls the mock
func (c *MockWebhookClient) GetWebhook(ctx context.Context, repo bitbucket.Repo, id int) (result bitbucket.Webhook, err error) {
	return c.MockGetWebhook(ctx, repo, id)
}

// UpdateWebhook calls the mock
func (c *MockWebhookClient) UpdateWebhook(ctx context.Context, repo bitbucket.Repo, id int, hook bitbucket.Webhook) (result bitbucket.Webhook, err error) {
	return c.MockUpdateWebhook(ctx, repo, id, hook)
}
