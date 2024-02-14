package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

var headerRegx *regexp.Regexp = regexp.MustCompile("[^a-zA-Z0-9-]+")

type Proxy struct {
	target     *url.URL
	targetPort int
	revp       *httputil.ReverseProxy
}

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	removeInvalidHeaders(&r.Header)
	proxy.revp.ServeHTTP(w, r)
}

func main() {
	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err == nil {
			port = p
		}
	}

	// This will be the port at which the real server will be running
	targetPort, err := getFreePort(port)
	if err != nil {
		panic("Can't get free port")
	}

	go spinRealServer(targetPort)

	remote, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))
	if err != nil {
		panic(err)
	}

	revp := httputil.NewSingleHostReverseProxy(remote)
	// Configure the reverse proxy to use HTTPS
	revp.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create a new Proxy instance
	proxy := &Proxy{
		target:     remote,
		targetPort: targetPort,
		revp:       revp,
	}

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), proxy)
	if err != nil {
		panic(err)
	}
}

func spinRealServer(port int) {
	// Check if there are additional arguments
	if len(os.Args) > 1 {
		cmd := exec.Command(os.Args[1], os.Args[2:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = cmd.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))

		// Start the specified command
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error starting command: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Started process %s with PID %d\n", os.Args[1], cmd.Process.Pid)
	}
}

func getFreePort(otherThan int) (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			port := l.Addr().(*net.TCPAddr).Port
			if port != otherThan {
				return port, nil
			} else {
				port, err := getFreePort(otherThan)
				return port, err
			}
		}
	}
	return
}

func removeInvalidHeaders(headers *http.Header) {
	for k, _ := range *headers {
		if headerRegx.MatchString(k) {
			headers.Del(k)
		}
	}
}
