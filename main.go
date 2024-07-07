package main

import "github.com/shivam-bhadani/distributed-cache/server"

func main() {
	server := server.NewServer(":8080")
	server.Start()
}
