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
	"compress/gzip"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
)

func encodeAndCompress(w io.Writer, store map[string][]byte) error {
	gz := gzip.NewWriter(w)
	defer gz.Close()
	return gob.NewEncoder(gz).Encode(store)
}

func decodeAndDecompress(r io.Reader, store *map[string][]byte) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()
	return gob.NewDecoder(gz).Decode(store)
}

func openStoreFileForWrite(path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
}

func openStoreFileForRead(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if info.Size() == 0 {
		f.Close()
		return nil, io.EOF
	}
	return f, nil
}

func (s *Server) PersistMemoryStore() error {
	path := s.config.MemoryDumpFileFullPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		s.logger.Error("Failed to create directory for memory store dump", "error", err.Error())
		return err
	}

	f, err := openStoreFileForWrite(path)
	if err != nil {
		s.logger.Error("Failed to open memory store dump file", "error", err.Error())
		return err
	}
	defer f.Close()

	if err := encodeAndCompress(f, s.store); err != nil {
		s.logger.Error("Failed to encode and compress the memory store", "error", err.Error())
		return err
	}

	s.logger.Info("Successfully written memory store dump to file", "path", path)
	return nil
}

func (s *Server) ReadPersistedMemoryStore() error {
	path := s.config.MemoryDumpFileFullPath()
	f, err := openStoreFileForRead(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, io.EOF) {
			s.logger.Warn("Memory store dump file does not exist or is empty, starting with empty store")
			return nil
		}
		s.logger.Error("Failed to open memory store dump file", "error", err.Error())
		return err
	}
	defer f.Close()

	if err := decodeAndDecompress(f, &s.store); err != nil {
		s.logger.Error("Failed to decode and decompress memory store dump file", "error", err.Error())
		return err
	}

	s.logger.Info("Successfully read memory store dump into memory", "size", len(s.store))
	return nil
}
