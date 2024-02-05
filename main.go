package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
)

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

var outPort int

func init() {
	var err error
	outPort, err = GetFreePort()
	if err != nil {
		panic("Can't get free port")
	}
	// Check if there are additional arguments
	if len(os.Args) > 1 {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", outPort))

		// Start the specified command
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error starting command: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Started process %s with PID %d\n", os.Args[1], cmd.Process.Pid)

		// Wait for the command to finish
		err = cmd.Wait()
		if err != nil {
			fmt.Printf("Command finished with error: %v\n", err)
		}

		// Exit after running the command
		os.Exit(0)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to port 8080 if no PORT env variable is set
	}
	// Start the proxy server on port 8080
	err := startProxy("127.0.0.1:" + port)
	if err != nil {
		panic(err)
	}
}

// ... (rest of the code remains the same)

func startProxy(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	fmt.Printf("Proxy server listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	clientReader := bufio.NewReader(clientConn)

	// Read the request from the client
	request, err := http.ReadRequest(clientReader)
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		return
	}

	// Filter invalid headers
	filterInvalidHeaders(request.Header)

	// Connect to the destination server
	outConn := fmt.Sprintf("localhost:%d", outPort)
	destConn, err := net.Dial("tcp", outConn)
	if err != nil {
		fmt.Printf("Error connecting to destination: %v\n", err)
		return
	}
	defer destConn.Close()

	// Write the modified request to the destination
	err = request.Write(destConn)
	if err != nil {
		fmt.Printf("Error writing request to destination: %v\n", err)
		return
	}

	// Copy the response from the destination to the client
	io.Copy(clientConn, destConn)
}

func filterInvalidHeaders(headers http.Header) {
	// Define invalid headers
	invalidHeaders := []string{"!~Passenger-Proto", "!~Passenger-Client-Address", "!~Passenger-Envvars"}

	for _, header := range invalidHeaders {
		if _, ok := headers[header]; ok {
			delete(headers, header)
		}
	}
}
