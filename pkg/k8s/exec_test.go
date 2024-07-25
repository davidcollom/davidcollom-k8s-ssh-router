package k8s

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/remotecommand"
)

func initTestCache() {
	fmt.Printf("InitCache Items(pre): %v\n", localCache.ItemCount())
	localCache.Set("default-testuser", map[string]string{
		"password":         "testpassword",
		"publicKey":        "testpublickey",
		"service":          "default",
		"podLabelSelector": "testpodlabelselector",
		"containerName":    "testcontainer",
		"shell":            "/bin/sh",
	}, cache.DefaultExpiration)
	fmt.Printf("InitCache Items(post): %v\n", localCache.ItemCount())
}

type mockChannel struct {
	ssh.Channel
	mock.Mock
}

func (m *mockChannel) Read(data []byte) (int, error) {
	return len(data), nil
}

func (m *mockChannel) Write(data []byte) (int, error) {
	return len(data), nil
}

func (m *mockChannel) Stderr() io.ReadWriter {
	return &bytes.Buffer{}
}

func (m *mockChannel) Close() error {
	return nil
}

type mockExecutor struct {
	StreamFunc func(options remotecommand.StreamOptions) error
}

func (m *mockExecutor) Stream(options remotecommand.StreamOptions) error {
	return m.StreamFunc(options)
}

func TestExecInPod(t *testing.T) {
	clientset := clientFake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels: map[string]string{
				"testpodlabelselector": "true",
			},
		},
	})

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

	executor := &mockExecutor{
		StreamFunc: func(options remotecommand.StreamOptions) error {
			return nil
		},
	}

	// Override the secret cache with the test cache
	localCache = cache.New(cache.NoExpiration, cache.NoExpiration)

	// Add user secret to the local cache
	initTestCache()

	channel := &mockChannel{}

	// Ensure that the cache is being checked correctly
	t.Logf("Cache Items: %v", localCache.ItemCount())
	userSecret, found := GetSecretFromCache("default-testuser")
	require.True(t, found, "User secret should be found in cache")
	require.NotNil(t, userSecret, "User secret should not be nil")

	err := ExecInPod(clientset, restClient, executor, config, "default-testuser", "echo hello", channel, false)
	require.NoError(t, err, "ExecInPod should not return an error")
}
