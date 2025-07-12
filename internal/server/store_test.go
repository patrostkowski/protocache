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

package server

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecodeStore(t *testing.T) {
	original := map[string][]byte{
		"foo": []byte("bar"),
		"baz": []byte("qux"),
	}

	var buf bytes.Buffer

	err := encodeAndCompress(&buf, original)
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)

	var decoded map[string][]byte
	err = decodeAndDecompress(&buf, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestOpenStoreFileForWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.gob.gz")

	w, err := openStoreFileForWrite(path)
	assert.NoError(t, err)
	defer w.Close()

	testData := map[string][]byte{"k": []byte("v")}
	err = encodeAndCompress(w, testData)
	assert.NoError(t, err)

	r, err := openStoreFileForRead(path)
	assert.NoError(t, err)
	defer r.Close()

	var decoded map[string][]byte
	err = decodeAndDecompress(r, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, testData, decoded)
}

func TestOpenStoreFileForRead_FileNotExist(t *testing.T) {
	_, err := openStoreFileForRead("nonexistent.gob.gz")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestOpenStoreFileForRead_EmptyFile(t *testing.T) {
	f, err := os.CreateTemp("", "empty-*.gob.gz")
	assert.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	r, err := openStoreFileForRead(f.Name())
	assert.ErrorIs(t, err, io.EOF)
	assert.Nil(t, r)
}
