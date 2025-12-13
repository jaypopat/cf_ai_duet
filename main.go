package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jaypopat/duet/internal/server"
)

func main() {
	addr := flag.String("addr", ":2222", "SSH server address")
	hostKeyPath := flag.String("hostkey", ".ssh/id_ed25519", "Path to SSH host key")
	flag.Parse()

	fmt.Println("Duet - SSH Pair Programming")
	fmt.Printf("Starting server on %s\n", *addr)

	srv := server.New(*addr, *hostKeyPath)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
