package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	pb "github.com/patrostkowski/protocache/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	host := flag.String("host", "localhost", "gRPC server host")
	port := flag.Int("port", 8080, "gRPC server port")
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		usage()
		return
	}

	target := fmt.Sprintf("%s:%d", *host, *port)

	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", target, err)
	}
	defer conn.Close()

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
  protocachecli [-host localhost] [-port 8080] <command> [args]

Commands:
  set <key> <value>     Set a value
  get <key>             Get a value
  del <key>             Delete a key
  clear                 Clear the cache`)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
