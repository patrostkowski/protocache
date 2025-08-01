# Copyright 2025 Patryk Rostkowski
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.24-bullseye AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o protocache ./cmd/protocache
RUN go build -o protocachecli ./cmd/protocachecli

FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/protocache /usr/local/bin/protocache
COPY --from=builder /app/protocachecli /usr/local/bin/protocachecli

ENTRYPOINT ["/usr/local/bin/protocache"]
