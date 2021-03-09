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

	"github.com/crossplane/provider-bitbucket-server/internal/clients/bitbucket"
)

func (c *Client) GetWebhook(ctx context.Context, repo bitbucket.Repo, id int) (bitbucket.Webhook, error) {
	url := c.BaseURL + fmt.Sprintf("/rest/1.0/projects/%s/repos/%s/webhooks/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	// The documentation says this is a paged API but it is not
	var payload KeyDescription
	if err := c.sendRequest(req, &payload); err != nil {
		return bitbucket.Webhook{}, fmt.Errorf("GetWebhook(%+v, %d): %w", repo, id, err)
	}

	return bitbucket.Webhook{
		// TODO
	}, nil
}

func (c *Client) CreateWebhook(ctx context.Context, repo bitbucket.Repo, key bitbucket.Webhook) (bitbucket.Webhook, error) {
	payload := UploadKeyPayload{
		// TODO
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	url := c.BaseURL + fmt.Sprintf("/rest/1.0/projects/%s/repos/%s/webhooks",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(marshalledPayload))
	if err != nil {
		return bitbucket.Webhook{}, err
	}

	var response KeyDescription
	if err := c.sendRequest(req, &response); err != nil {
		return bitbucket.Webhook{}, err
	}
	return bitbucket.Webhook{
		// TODO
	}, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, repo bitbucket.Repo, id int) error {
	url := c.BaseURL + fmt.Sprintf("/rest/1.0/projects/%s/repos/%s/webhooks/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}
