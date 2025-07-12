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

package config

import (
	"fmt"
	"time"
)

const (
	GRPCPort              = 50051
	HTTPPort              = 9091
	ListenAddr            = "0.0.0.0"
	ServerShutdownTimeout = 30 * time.Second
	GracefulTimeout       = 10 * time.Second
	MemoryDumpPath        = "/var/lib/protocache/"
	MemoryDumpFileName    = "protocache.gob.gz"
)

var (
	GRPCAddr = fmt.Sprintf("%s:%d", ListenAddr, GRPCPort)
	HTTPAddr = fmt.Sprintf("%s:%d", ListenAddr, HTTPPort)

	MemoryDumpFileFullPath = MemoryDumpPath + MemoryDumpFileName
)

type Config struct {
	MemoryDumpFilePath string
}

func DefaultConfig() *Config {
	return &Config{
		MemoryDumpFilePath: MemoryDumpFileFullPath,
	}
}
