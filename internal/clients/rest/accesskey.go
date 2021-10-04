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

// ListAccessKeys returns all access keys for the given repository
func (c *Client) ListAccessKeys(ctx context.Context, repo bitbucket.Repo) ([]bitbucket.AccessKey, error) {
	url := c.BaseURL + fmt.Sprintf("/rest/keys/1.0/projects/%s/repos/%s/ssh",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var payload GetKeysPayload
	if err := c.sendRequest(req, &payload); err != nil {
		return nil, fmt.Errorf("ListAccessKeys(%+v): %w", repo, err)
	}

	ret := make([]bitbucket.AccessKey, 0, len(payload.Values))
	for _, key := range payload.Values {
		ret = append(ret, bitbucket.AccessKey{
			Key:        key.Key.Text,
			ID:         key.Key.ID,
			Label:      key.Key.Label,
			Permission: key.Permission,
		})
	}

	return ret, nil
}

// GetAccessKey finds accesskey given by bitbucket server access key id
func (c *Client) GetAccessKey(ctx context.Context, repo bitbucket.Repo, id int) (bitbucket.AccessKey, error) {
	url := c.BaseURL + fmt.Sprintf("/rest/keys/1.0/projects/%s/repos/%s/ssh/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return bitbucket.AccessKey{}, err
	}

	// The documentation says this is a paged API but it is not
	var payload KeyDescription
	if err := c.sendRequest(req, &payload); err != nil {
		return bitbucket.AccessKey{}, fmt.Errorf("GetAccessKey(%+v, %d): %w", repo, id, err)
	}

	return bitbucket.AccessKey{
		Key:        payload.Key.Text,
		ID:         payload.Key.ID,
		Label:      payload.Key.Label,
		Permission: payload.Permission,
	}, nil
}

// CreateAccessKey on a repository by providing the public key
func (c *Client) CreateAccessKey(ctx context.Context, repo bitbucket.Repo, key bitbucket.AccessKey) (bitbucket.AccessKey, error) {
	payload := UploadKeyPayload{
		Key: PublicSSHKey{
			Text:  key.Key,
			Label: key.Label,
		},
		Permission: key.Permission,
	}

	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return bitbucket.AccessKey{}, err
	}

	url := c.BaseURL + fmt.Sprintf("/rest/keys/1.0/projects/%s/repos/%s/ssh",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(marshalledPayload))
	if err != nil {
		return bitbucket.AccessKey{}, err
	}

	var response KeyDescription
	if err := c.sendRequest(req, &response); err != nil {
		return bitbucket.AccessKey{}, err
	}
	return bitbucket.AccessKey{
		ID:         response.Key.ID,
		Key:        response.Key.Text,
		Label:      response.Key.Label,
		Permission: response.Permission,
	}, nil
}

// UpdateAccessKeyPermission enables mutation of permissions on a accesskey by providing the id of the access key.
// Valid options are REPO_READ or REPO_WRITE
func (c *Client) UpdateAccessKeyPermission(ctx context.Context, repo bitbucket.Repo, id int, permission string) error {
	url := c.BaseURL + fmt.Sprintf("/rest/keys/1.0/projects/%s/repos/%s/ssh/%d/permission/%s",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id, permission)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}

// DeleteAccessKey removes the key from the repository by provided access key id
func (c *Client) DeleteAccessKey(ctx context.Context, repo bitbucket.Repo, id int) error {
	url := c.BaseURL + fmt.Sprintf("/rest/keys/1.0/projects/%s/repos/%s/ssh/%d",
		url.PathEscape(repo.ProjectKey), url.PathEscape(repo.Repo), id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}

// PublicSSHKey represents the public ssh key
type PublicSSHKey struct {
	// Text contains the public key
	Text string `json:"text"`
	// Labels describes the public key
	Label string `json:"label"`
}

// UploadKeyPayload defines api object for key upload
type UploadKeyPayload struct {
	// Key defines the type of public ssh key
	Key PublicSSHKey `json:"key"`
	// Permissions defines the access level for the access key in bitbucket server
	Permission string `json:"permission"`
}

// GetKeysPayload is the returned object from bitbucket server
type GetKeysPayload struct {
	// Pagination is defined by the bitbucket server api
	Pagination `json:",inline"`
	// Values is defined by the bitbucket server api
	Values []KeyDescription `json:"values"`
}

// KeyDescription describes a specific accesskey in bitbucket server
type KeyDescription struct {
	// Key contains info about the access key
	Key KeyInfo `json:"key"`
	// Repository contains information about the repository where the access key is added
	Repository RepositoryInfo `json:"repository"`
	// Permission is the level of permission the access key has been granted
	Permission string `json:"permission"`
}

// KeyInfo contains the information about the access key
type KeyInfo struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Label string `json:"label"`
}

// RepositoryInfo contains information about the repository
type RepositoryInfo struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	Project ProjectInfo
}

// ProjectInfo contains information on the project
type ProjectInfo struct {
	Key string `json:"key"`
}
