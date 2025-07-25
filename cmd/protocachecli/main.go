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
	"net"
	"time"
	"unicode/utf8"

	pb "github.com/patrostkowski/protocache/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	host := flag.String("host", "localhost", "gRPC server host")
	port := flag.Int("port", 50051, "gRPC server port")
	socket := flag.String("socket", "", "Unix socket path (overrides host/port)")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		usage()
		return
	}

	var (
		conn *grpc.ClientConn
		err  error
	)
	if *socket != "" {
		conn, err = grpc.NewClient(
			"unix://"+*socket,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return net.Dial("unix", *socket)
			}),
		)
	} else {
		target := fmt.Sprintf("%s:%d", *host, *port)
		conn, err = grpc.NewClient(target,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	defer conn.Close()

	if err != nil {
		panic(err)
	}

	client := pb.NewCacheServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := args[0]
	params := args[1:]

	switch cmd {
	case "set":
		if len(params) != 2 {
			fmt.Println("Usage: set <key> <value>")
			return
		}
		_, err := client.Set(ctx, &pb.SetRequest{Key: params[0], Value: []byte(params[1])})
		checkErr(err)
		fmt.Println("OK")

	case "get":
		if len(params) != 1 {
			fmt.Println("Usage: get <key>")
			return
		}
		res, err := client.Get(ctx, &pb.GetRequest{Key: params[0]})
		checkErr(err)
		if !res.Found {
			fmt.Println("(nil)")
		} else if utf8.Valid(res.Value) {
			fmt.Printf("%s\n", res.Value)
		} else {
			fmt.Printf("(binary) %s\n", base64.StdEncoding.EncodeToString(res.Value))
		}

	case "del":
		if len(params) != 1 {
			fmt.Println("Usage: del <key>")
			return
		}
		_, err := client.Delete(ctx, &pb.DeleteRequest{Key: params[0]})
		checkErr(err)
		fmt.Println("Deleted")

	case "clear":
		_, err := client.Clear(ctx, &pb.ClearRequest{})
		checkErr(err)
		fmt.Println("Cache cleared")

	case "list":
		res, err := client.List(ctx, &pb.ListRequest{})
		checkErr(err)
		for _, k := range res.Keys {
			fmt.Println(k)
		}

	case "stats":
		res, err := client.Stats(ctx, &pb.StatsRequest{})
		checkErr(err)
		fmt.Println(res)

	case "help":
		usage()

	default:
		fmt.Println("Unknown command:", cmd)
		usage()
	}
}

func usage() {
	fmt.Println(`Usage:
  protocachecli [-host localhost] [-port 50051] [-socket /path/to/socket] <command> [args]

Flags:
  -host    gRPC TCP hostname (ignored if -socket is set)
  -port    gRPC TCP port (ignored if -socket is set)
  -socket  Path to Unix socket (takes priority over host:port)
  
Commands:
  set <key> <value>     Set a value
  get <key>             Get a value
  del <key>             Delete a key
  list                  List all keys
  stats                 Print server stats
  clear                 Clear the cache`)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
