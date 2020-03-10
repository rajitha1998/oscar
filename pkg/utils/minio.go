// Copyright (C) GRyCAP - I3M - UPV
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/grycap/oscar/pkg/types"
	"github.com/minio/minio/pkg/madmin"
)

// MinIOAdminClient struct to represent a MinIO Admin client to configure webhook notifications
type MinIOAdminClient struct {
	adminClient   *madmin.AdminClient
	oscarEndpoint *url.URL
}

// MakeMinIOAdminClient creates a new MinIO Admin client to configure webhook notifications
func MakeMinIOAdminClient(provider *types.MinIOProvider, cfg *types.Config) (*MinIOAdminClient, error) {
	// Check if both MinIOProvider and Config minIO endpoints are the same
	// TODO: allow registering webhooks on external minIO servers
	// Functionality to retrieve the oscar service/loadbalancer external IP and port/nodeport is required to support it
	if provider.Endpoint.String() != cfg.MinIOEndpoint.String() {
		return nil, fmt.Errorf("The provided MinIO server \"%s\" is not the same as the one configured", provider.Endpoint.String())
	}

	// Check URL Scheme for using TLS or not
	var enableTLS bool
	switch provider.Endpoint.Scheme {
	case "http":
		enableTLS = false
	case "https":
		enableTLS = true
	default:
		return nil, fmt.Errorf("Invalid MinIO Endpoint: %s", provider.Endpoint.String())
	}

	adminClient, err := madmin.New(provider.Endpoint.Host, provider.AccessKey, provider.SecretKey, enableTLS)
	if err != nil {
		return nil, err
	}

	// Disable tls verification in client transport if verify == false
	if !provider.Verify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		adminClient.SetCustomTransport(tr)
	}

	oscarEndpoint, err := url.Parse(fmt.Sprintf("http://%s.%s:%d", cfg.Name, cfg.Namespace, cfg.ServicePort))
	if err != nil {
		return nil, err
	}

	minIOAdminClient := &MinIOAdminClient{
		adminClient:   adminClient,
		oscarEndpoint: oscarEndpoint,
	}

	return minIOAdminClient, nil
}

// RegisterMinIOWebhook registers a new webhook in the MinIO configuration
func (minIOAdminClient *MinIOAdminClient) RegisterMinIOWebhook(name string) error {
	err := minIOAdminClient.adminClient.SetConfigKV(fmt.Sprintf("notify_webhook:%s endpoint=%s/job/%s", name, minIOAdminClient.oscarEndpoint.String(), name))
	if err != nil {
		return err
	}
	return nil
}

// RemoveMinIOWebhook removes an existent webhook in the MinIO configuration
func (minIOAdminClient *MinIOAdminClient) RemoveMinIOWebhook(name string) error {
	err := minIOAdminClient.adminClient.DelConfigKV(fmt.Sprintf("notify_webhook:%s", name))
	if err != nil {
		return err
	}
	return nil
}

// RestartMinIOServer restarts a MinIO server to apply the configuration changes
func (minIOAdminClient *MinIOAdminClient) RestartMinIOServer() error {
	err := minIOAdminClient.adminClient.ServiceRestart()
	if err != nil {
		return err
	}

	// Max. time taken by the server to shutdown is 5 seconds.
	// This can happen when there are lot of s3 requests pending when the server
	// receives a restart command.
	// Sleep for 6 seconds and then check if the server is online.
	time.Sleep(6 * time.Second)
	_, err = minIOAdminClient.adminClient.ServerInfo()
	if err != nil {
		return fmt.Errorf("Error restarting the MinIO server: %v", err)
	}

	return nil
}