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

var _ bitbucket.KeyClientAPI = &MockKeyClient{}

// MockKeyClient is a fake implementation of KeyClientAPI
type MockKeyClient struct {
	bitbucket.KeyClientAPI

	MockCreateAccessKey           func(ctx context.Context, repo bitbucket.Repo, key bitbucket.AccessKey) (result bitbucket.AccessKey, err error)
	MockDeleteAccessKey           func(ctx context.Context, repo bitbucket.Repo, id int) (err error)
	MockGetAccessKey              func(ctx context.Context, repo bitbucket.Repo, id int) (result bitbucket.AccessKey, err error)
	MockListAccessKeys            func(ctx context.Context, repo bitbucket.Repo) (result []bitbucket.AccessKey, err error)
	MockUpdateAccessKeyPermission func(ctx context.Context, repo bitbucket.Repo, id int, permission string) (err error)
}

// CreateAccessKey calls the mock
func (c *MockKeyClient) CreateAccessKey(ctx context.Context, repo bitbucket.Repo, key bitbucket.AccessKey) (result bitbucket.AccessKey, err error) {
	return c.MockCreateAccessKey(ctx, repo, key)
}

// DeleteAccessKey calls the mock
func (c *MockKeyClient) DeleteAccessKey(ctx context.Context, repo bitbucket.Repo, id int) (err error) {
	return c.MockDeleteAccessKey(ctx, repo, id)
}

// GetAccessKey calls the mock
func (c *MockKeyClient) GetAccessKey(ctx context.Context, repo bitbucket.Repo, id int) (result bitbucket.AccessKey, err error) {
	return c.MockGetAccessKey(ctx, repo, id)
}

// ListAccessKeys calls the mock
func (c *MockKeyClient) ListAccessKeys(ctx context.Context, repo bitbucket.Repo) (result []bitbucket.AccessKey, err error) {
	return c.MockListAccessKeys(ctx, repo)
}

// UpdateAccessKeyPermission calls the mock
func (c *MockKeyClient) UpdateAccessKeyPermission(ctx context.Context, repo bitbucket.Repo, id int, permission string) error {
	return c.MockUpdateAccessKeyPermission(ctx, repo, id, permission)
}
