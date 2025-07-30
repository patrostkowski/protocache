// Copyright 2025 Patryk Rostkowski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Host    string
	Port    int
	Socket  string
	Cert    string
	Key     string
	Timeout time.Duration
}

type Client struct {
	conn   *grpc.ClientConn
	client cachev1alpha.CacheServiceClient
}

func New(cfg Config) (*Client, error) {
	var opts []grpc.DialOption

	if cfg.Cert != "" && cfg.Key != "" {
		certificate, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			return nil, fmt.Errorf("loading TLS cert/key: %w", err)
		}
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{certificate},
			InsecureSkipVerify: true, // #nosec G402
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	var (
		conn *grpc.ClientConn
		err  error
	)
	if cfg.Socket != "" {
		opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", cfg.Socket)
		}))
		conn, err = grpc.NewClient("unix://"+cfg.Socket, opts...)
	} else {
		target := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		conn, err = grpc.NewClient(target, opts...)
	}

	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: cachev1alpha.NewCacheServiceClient(conn),
	}, nil
}

// Close gracefully closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Set stores a key-value pair in the cache.
func (c *Client) Set(ctx context.Context, key, value string) error {
	_, err := c.client.Set(ctx, &cachev1alpha.SetRequest{
		Key:   key,
		Value: []byte(value),
	})
	return err
}

// Get retrieves a value from the cache by key.
func (c *Client) Get(ctx context.Context, key string) (*cachev1alpha.GetResponse, error) {
	return c.client.Get(ctx, &cachev1alpha.GetRequest{Key: key})
}

// Delete removes a key from the cache.
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, &cachev1alpha.DeleteRequest{Key: key})
	return err
}

// Clear removes all keys from the cache.
func (c *Client) Clear(ctx context.Context) error {
	_, err := c.client.Clear(ctx, &cachev1alpha.ClearRequest{})
	return err
}

// List returns a list of all keys currently in the cache.
func (c *Client) List(ctx context.Context) ([]string, error) {
	res, err := c.client.List(ctx, &cachev1alpha.ListRequest{})
	if err != nil {
		return nil, err
	}
	return res.Keys, nil
}

// Stats retrieves server statistics.
func (c *Client) Stats(ctx context.Context) (*cachev1alpha.StatsResponse, error) {
	return c.client.Stats(ctx, &cachev1alpha.StatsRequest{})
}
