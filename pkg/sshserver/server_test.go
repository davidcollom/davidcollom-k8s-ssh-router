package sshserver

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	clientFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
)

func TestStartSSHServer(t *testing.T) {
	privateBytes, err := generatePrivateKey()
	require.NoError(t, err, "Failed to generate private key")

	tmpDir := t.TempDir()
	privateKeyPath := filepath.Join(tmpDir, "id_rsa")
	err = os.WriteFile(privateKeyPath, privateBytes, 0600)
	require.NoError(t, err, "Failed to write private key to file")

	clientset := clientFake.NewSimpleClientset()
	restClient := &fake.RESTClient{
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       http.NoBody,
			}, nil
		}),
	}
	config := &rest.Config{Host: "http://localhost"}

	port, err := getRandomPort()
	require.NoError(t, err, "Failed to get a random port")

	listenAddr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	ln, err := net.Listen("tcp", listenAddr)
	require.NoError(t, err, "Failed to listen on %s", listenAddr)

	go func() {
		conn, err := ln.Accept()
		require.NoError(t, err, "Failed to accept connection")

		signer, err := ssh.ParsePrivateKey(privateBytes)
		require.NoError(t, err, "Failed to parse private key")

		serverConfig := &ssh.ServerConfig{
			NoClientAuth: true,
		}
		serverConfig.AddHostKey(signer)

		sshConn, chans, reqs, err := ssh.NewServerConn(conn, serverConfig)
		require.NoError(t, err, "Failed to establish SSH connection")

		go ssh.DiscardRequests(reqs)
		for newChannel := range chans {
			channel, requests, err := newChannel.Accept()
			require.NoError(t, err, "Failed to accept channel")

			go handleSSHRequests(clientset, restClient, config, nil, channel, requests, sshConn.User())
		}
	}()

	conn, err := net.Dial("tcp", listenAddr)
	require.NoError(t, err, "Failed to dial %s", listenAddr)

	clientConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshClientConn, chans, reqs, err := ssh.NewClientConn(conn, listenAddr, clientConfig)
	require.NoError(t, err, "Failed to establish SSH client connection")
	defer sshClientConn.Close()

	go ssh.DiscardRequests(reqs)
	go func() {
		for newChannel := range chans {
			channel, requests, err := newChannel.Accept()
			require.NoError(t, err, "Failed to accept channel")

			go func() {
				for req := range requests {
					if req.WantReply {
						req.Reply(true, nil)
					}
				}
			}()
			channel.Close()
		}
	}()

	time.Sleep(1 * time.Second) // Give some time for the connection handling
}
