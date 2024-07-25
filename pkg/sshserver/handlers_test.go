package sshserver

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/davidcollom/k8s-ssh-router/pkg/k8s"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	clientFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/remotecommand"
)

func initTestCache() {
	k8s.SetSecretInCache("default-testuser", map[string]string{
		"password":         "testpassword",
		"publicKey":        "testpublickey",
		"service":          "default",
		"podLabelSelector": "testpodlabelselector",
		"containerName":    "testcontainer",
		"shell":            "/bin/sh",
	})
}

type mockChannel struct {
	ssh.Channel
	mock.Mock
}

func (m *mockChannel) Read(data []byte) (int, error) {
	args := m.Called(data)
	return args.Int(0), args.Error(1)
}

func (m *mockChannel) Write(data []byte) (int, error) {
	args := m.Called(data)
	return args.Int(0), args.Error(1)
}

func (m *mockChannel) Stderr() io.ReadWriter {
	return &bytes.Buffer{}
}

func (m *mockChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockExecutor struct {
	StreamFunc func(options remotecommand.StreamOptions) error
}

func (m *mockExecutor) Stream(options remotecommand.StreamOptions) error {
	return m.StreamFunc(options)
}

type mockSSHRequest struct {
	ssh.Request
	mock.Mock
}

func (m *mockSSHRequest) Reply(ok bool, payload []byte) error {
	args := m.Called(ok, payload)
	return args.Error(0)
}

func TestHandleSSHRequests(t *testing.T) {
	clientset := clientFake.NewSimpleClientset()
	config := &rest.Config{Host: "http://localhost"}

	executor := &mockExecutor{
		StreamFunc: func(options remotecommand.StreamOptions) error {
			return nil
		},
	}

	initTestCache()

	channel := &mockChannel{}
	channel.On("Close").Return(nil) // Mocking the Close method

	restClient := &fake.RESTClient{
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       http.NoBody,
			}, nil
		}),
	}

	t.Run("exec request", func(t *testing.T) {
		reqs := make(chan *ssh.Request, 1)
		reqs <- &ssh.Request{
			Type:      "exec",
			WantReply: true,
			Payload:   []byte{0, 0, 0, 0, 'e', 'c', 'h', 'o', ' ', 'h', 'e', 'l', 'l', 'o'},
		}
		close(reqs)

		handleSSHRequests(clientset, restClient, config, executor, channel, reqs, "default-testuser")
	})

	t.Run("shell request", func(t *testing.T) {
		reqs := make(chan *ssh.Request, 1)
		reqs <- &ssh.Request{
			Type:      "shell",
			WantReply: true,
		}
		close(reqs)

		handleSSHRequests(clientset, restClient, config, executor, channel, reqs, "default-testuser")
	})

	// t.Run("pty-req request", func(t *testing.T) {
	// 	reqs := make(chan *ssh.Request, 1)
	// 	req := &mockSSHRequest{}
	// 	req.On("Reply", true, []byte(nil)).Return(nil)
	// 	req.Type = "pty-req"
	// 	req.WantReply = true
	// 	req.Payload = []byte{0, 0, 0, 0, 80, 24, 0, 0, 0, 0, 0, 0, 0} // Payload for pty-req with width and height
	// 	reqs <- &req.Request
	// 	close(reqs)

	// 	handleSSHRequests(clientset, restClient, config, executor, channel, reqs, "default-testuser")

	// 	req.AssertExpectations(t)
	// })

	// t.Run("subsystem request", func(t *testing.T) {
	// 	reqs := make(chan *ssh.Request, 1)
	// 	reqs <- &ssh.Request{
	// 		Type:      "subsystem",
	// 		WantReply: true,
	// 		Payload:   []byte("sftp"),
	// 	}
	// 	close(reqs)

	// 	handleSSHRequests(clientset, restClient, config, executor, channel, reqs, "default-testuser")
	// })

	// t.Run("unknown request", func(t *testing.T) {
	// 	reqs := make(chan *ssh.Request, 1)
	// 	reqs <- &ssh.Request{
	// 		Type:      "unknown",
	// 		WantReply: true,
	// 	}
	// 	close(reqs)

	// 	handleSSHRequests(clientset, restClient, config, executor, channel, reqs, "default-testuser")
	// })
}

func generatePrivateKey() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := new(bytes.Buffer)
	err = pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, err
	}

	return privateKeyPEM.Bytes(), nil
}

func getRandomPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func TestHandleSSHConnection(t *testing.T) {
	privateBytes, err := generatePrivateKey()
	require.NoError(t, err, "Failed to generate private key")

	signer, err := ssh.ParsePrivateKey(privateBytes)
	require.NoError(t, err, "Failed to parse private key")

	serverConfig := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	serverConfig.AddHostKey(signer)

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
