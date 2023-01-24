package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"golang.org/x/term"
)

var isServer *bool

func init() {
	isServer = flag.Bool("server", false, "")
}

func server() error {

	// Create command
	c := exec.Command(os.Getenv("SHELL"))

	// Start the command with a pty.
	ptmx, e := pty.Start(c)
	if e != nil {
		return e
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }()

	return listen(ptmx)
}

func listen(ptmx *os.File) error {
	fmt.Println("Launching server on 0.0.0.0:8088")

	// listen on all interfaces
	ln, e := net.Listen("tcp", "0.0.0.0:8088")
	if e != nil {
		return e
	}

	// accept connection on port
	conn, e := ln.Accept()
	if e != nil {
		return e
	}
	fmt.Println("The client connection was accepted.")

	go func() { _, _ = io.Copy(ptmx, conn) }()

	_, e = io.Copy(conn, ptmx)

	return e
}

func client() error {
	// connect to this socket
	conn, e := net.Dial("tcp", "127.0.0.1:8088")
	if e != nil {
		return e
	}

	// MakeRaw put the terminal connected to the given file descriptor into raw
	// mode and returns the previous state of the terminal so that it can be
	// restored.
	oldState, e := term.MakeRaw(int(os.Stdin.Fd()))
	if e != nil {
		return e
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() { _, _ = io.Copy(os.Stdout, conn) }()
	_, e = io.Copy(conn, os.Stdin)

	fmt.Println("Bye!")

	return e
}

func clientAndServer() error {
	flag.Parse()

	// If runs the app with --server flag
	if isServer != nil && *isServer {
		fmt.Println("Starting server mode")
		return server()

	} else {
		fmt.Println("Starting client mode")
		return client()
	}
}

func main() {
	if e := clientAndServer(); e != nil {
		fmt.Println(e)
	}
}
