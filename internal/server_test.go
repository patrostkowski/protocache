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

package internal_test

import (
	"context"
	"testing"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGet(t *testing.T) {
	server := internal.NewServer()
	ctx := context.Background()

	_, err := server.Set(ctx, &pb.SetRequest{Key: "foo", Value: []byte("bar")})
	assert.NoError(t, err)

	res, err := server.Get(ctx, &pb.GetRequest{Key: "foo"})
	assert.NoError(t, err)
	assert.True(t, res.Found)
	assert.Equal(t, []byte("bar"), res.Value)

	res, err = server.Get(ctx, &pb.GetRequest{Key: "baz"})
	assert.NoError(t, err)
	assert.False(t, res.Found)
}

func TestDelete(t *testing.T) {
	server := internal.NewServer()
	ctx := context.Background()

	server.Set(ctx, &pb.SetRequest{Key: "foo", Value: []byte("bar")})

	_, err := server.Delete(ctx, &pb.DeleteRequest{Key: "foo"})
	assert.NoError(t, err)

	res, _ := server.Get(ctx, &pb.GetRequest{Key: "foo"})
	assert.False(t, res.Found)
}

func TestClear(t *testing.T) {
	server := internal.NewServer()
	ctx := context.Background()

	server.Set(ctx, &pb.SetRequest{Key: "a", Value: []byte("1")})
	server.Set(ctx, &pb.SetRequest{Key: "b", Value: []byte("2")})

	_, err := server.Clear(ctx, &pb.ClearRequest{})
	assert.NoError(t, err)

	resA, _ := server.Get(ctx, &pb.GetRequest{Key: "a"})
	resB, _ := server.Get(ctx, &pb.GetRequest{Key: "b"})
	assert.False(t, resA.Found)
	assert.False(t, resB.Found)
}
