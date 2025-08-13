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

package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	"github.com/patrostkowski/protocache/pkg/client"
)

func main() {
	cfg, cmd, params := parseFlags()

	c, err := client.New(cfg)
	checkErr(err)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	runCommand(ctx, c, cmd, params)
}

func parseFlags() (client.Config, string, []string) {
	host := flag.String("host", "localhost", "gRPC server host")
	port := flag.Int("port", 50051, "gRPC server port")
	socket := flag.String("socket", "", "Unix socket path (overrides host/port)")
	cert := flag.String("cert", "", "TLS client certificate file (enables TLS if set)")
	key := flag.String("key", "", "TLS client key file (requires --cert)")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		usage()
		log.Fatal("no command provided")
	}

	cfg := client.Config{
		Host:    *host,
		Port:    *port,
		Socket:  *socket,
		Cert:    *cert,
		Key:     *key,
		Timeout: 2 * time.Second,
	}

	return cfg, args[0], args[1:]
}

func runCommand(ctx context.Context, c *client.Client, cmd string, params []string) {
	switch cmd {
	case "set":
		runSet(ctx, c, params)
	case "get":
		runGet(ctx, c, params)
	case "del":
		runDelete(ctx, c, params)
	case "clear":
		runClear(ctx, c)
	case "list":
		runList(ctx, c)
	case "stats":
		runStats(ctx, c)
	case "help":
		usage()
	default:
		fmt.Println("Unknown command:", cmd)
		usage()
	}
}

func runSet(ctx context.Context, c *client.Client, params []string) {
	if len(params) != 2 {
		fmt.Println("Usage: set <key> <value>")
		return
	}
	err := c.Set(ctx, params[0], params[1])
	checkErr(err)
	fmt.Println("OK")
}

func runGet(ctx context.Context, c *client.Client, params []string) {
	if len(params) != 1 {
		fmt.Println("Usage: get <key>")
		return
	}
	res, err := c.Get(ctx, params[0])
	checkErr(err)
	if !res.Found {
		fmt.Println("(nil)")
	} else if utf8.Valid(res.Value) {
		fmt.Printf("%s\n", res.Value)
	} else {
		fmt.Printf("(binary) %s\n", base64.StdEncoding.EncodeToString(res.Value))
	}
}

func runDelete(ctx context.Context, c *client.Client, params []string) {
	if len(params) != 1 {
		fmt.Println("Usage: del <key>")
		return
	}
	err := c.Delete(ctx, params[0])
	checkErr(err)
	fmt.Println("Deleted")
}

func runClear(ctx context.Context, c *client.Client) {
	err := c.Clear(ctx)
	checkErr(err)
	fmt.Println("Cache cleared")
}

func runList(ctx context.Context, c *client.Client) {
	keys, err := c.List(ctx)
	checkErr(err)
	for _, k := range keys {
		fmt.Println(k)
	}
}

func runStats(ctx context.Context, c *client.Client) {
	stats, err := c.Stats(ctx)
	checkErr(err)
	fmt.Println(stats)
}

func usage() {
	fmt.Println(`Usage:
  protocachecli [-host localhost] [-port 50051] [-socket /path/to/socket] <command> [args]

Flags:
  -host    gRPC TCP hostname (ignored if -socket is set)
  -port    gRPC TCP port (ignored if -socket is set)
  -socket  Path to Unix socket (takes priority over host:port)
  -cert    TLS client certificate file (optional, enables TLS if set)
  -key     TLS client key file (required if --cert is set)

Commands:
  set <key> <value>     Set a value
  get <key>             Get a value
  del <key>             Delete a key
  list                  List all keys
  stats                 Print server stats
  clear                 Clear the cache
  help                  Show this help message`)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
