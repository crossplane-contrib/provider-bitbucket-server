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

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/crossplane-contrib/provider-bitbucket-server/internal/clients/bitbucket"
)

// GetWebhook gets the web hook
func (c *Client) GetWebhook(ctx context.Context, repo bitbucket.Repo, id int) (bitbucket.Webhook, error) {
	url := c.BaseURL + fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/webhooks/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	// The documentation says this is a paged API but it is not
	var payload bitbucket.Webhook
	if err := c.sendRequest(req, &payload); err != nil {
		return bitbucket.Webhook{}, fmt.Errorf("GetWebhook(%+v, %d): %w", repo, id, err)
	}

	return payload, nil
}

// CreateWebhook creates the web hook
func (c *Client) CreateWebhook(ctx context.Context, repo bitbucket.Repo, hook bitbucket.Webhook) (bitbucket.Webhook, error) {
	marshalledPayload, err := json.Marshal(hook)
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	url := c.BaseURL + fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/webhooks",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(marshalledPayload))
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	var response bitbucket.Webhook
	if err := c.sendRequest(req, &response); err != nil {
		return bitbucket.Webhook{}, err
	}
	return response, nil
}

// UpdateWebhook updates the web hook
func (c *Client) UpdateWebhook(ctx context.Context, repo bitbucket.Repo, id int, hook bitbucket.Webhook) (bitbucket.Webhook, error) {
	marshalledPayload, err := json.Marshal(hook)
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	url := c.BaseURL + fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/webhooks/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(marshalledPayload))
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	var response bitbucket.Webhook
	if err := c.sendRequest(req, &response); err != nil {
		return bitbucket.Webhook{}, err
	}
	return response, nil
}

// DeleteWebhook deletes the web hook
func (c *Client) DeleteWebhook(ctx context.Context, repo bitbucket.Repo, id int) error {
	url := c.BaseURL + fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/webhooks/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}
