package k8s

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informers "k8s.io/client-go/informers"
	clientFake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestProcessSecret(t *testing.T) {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("testuser"),
			"password": []byte("testpassword"),
		},
	}

	// Process the secret
	processSecret(secret)

	// Verify secret in cache
	usernameWithNamespace := "default-testuser"
	cachedSecret, found := GetSecretFromCache(usernameWithNamespace)
	assert.True(t, found, "Secret should be found in cache")
	expectedData := map[string]string{
		"password":         "testpassword",
		"publicKey":        "",
		"service":          "",
		"podLabelSelector": "",
		"containerName":    "",
		"shell":            "",
	}
	assert.Equal(t, expectedData, cachedSecret, "Secret data should match")
}

func TestWatchSecretsClusterWide(t *testing.T) {
	clientset := clientFake.NewSimpleClientset()

	// Create a channel to signal when the informer is ready
	readyCh := make(chan struct{})

	// Create the informer factory
	factory := informers.NewSharedInformerFactory(clientset, 0)

	// Create the informer
	informer := factory.Core().V1().Secrets().Informer()

	// Add event handlers to the informer
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Processing added event")
			secret := obj.(*v1.Secret)
			processSecret(secret)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Println("Processing modified event")
			secret := newObj.(*v1.Secret)
			processSecret(secret)
		},
		DeleteFunc: func(obj interface{}) {
			log.Println("Processing deleted event")
			secret := obj.(*v1.Secret)
			namespace := secret.Namespace
			username := string(secret.Data["username"])
			usernameWithNamespace := fmt.Sprintf("%s-%s", namespace, username)
			localCache.Delete(usernameWithNamespace)
			log.Printf("Deleted secret: %s, cache size: %d\n", usernameWithNamespace, localCache.ItemCount())
		},
	})

	// Start the informer
	go informer.Run(readyCh)

	if !cache.WaitForCacheSync(readyCh, informer.HasSynced) {
		log.Fatalf("failed to sync cache")
		t.Fatalf("failed to sync cache")
	}

	t.Log("Ready, LET'S GO!")

	// Create a secret to use in the test
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("testuser"),
			"password": []byte("testpassword"),
		},
	}

	// Add the secret
	_, err := clientset.CoreV1().Secrets("default").Create(context.TODO(), secret, metav1.CreateOptions{})
	require.NoError(t, err, "Failed to create test secret")

	// Wait for the cache to be updated
	cacheUpdated := make(chan struct{})
	go func() {
		for {
			if _, found := GetSecretFromCache("default-testuser"); found {
				cacheUpdated <- struct{}{}
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	select {
	case <-cacheUpdated:
		t.Log("Cache update verified")
	case <-time.After(10 * time.Second):
		t.Error("Timed out waiting for add event to be processed")
	}

	// Test for secret modification
	secret.Data["password"] = []byte("newpassword")
	_, err = clientset.CoreV1().Secrets("default").Update(context.TODO(), secret, metav1.UpdateOptions{})
	require.NoError(t, err, "Failed to update test secret")

	// Wait for the cache to be updated after modification
	cacheUpdated = make(chan struct{})
	go func() {
		for {
			if secretData, found := GetSecretFromCache("default-testuser"); found {
				if secretData.(map[string]string)["password"] == "newpassword" {
					cacheUpdated <- struct{}{}
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	select {
	case <-cacheUpdated:
		t.Log("Cache update verified for modification")
	case <-time.After(10 * time.Second):
		t.Error("Timed out waiting for modify event to be processed")
	}

	// Test for secret deletion
	err = clientset.CoreV1().Secrets("default").Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
	require.NoError(t, err, "Failed to delete test secret")

	// Wait for the cache to be updated after deletion
	cacheUpdated = make(chan struct{})
	go func() {
		for {
			if _, found := GetSecretFromCache("default-testuser"); !found {
				cacheUpdated <- struct{}{}
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	select {
	case <-cacheUpdated:
		t.Log("Cache update verified for deletion")
	case <-time.After(10 * time.Second):
		t.Error("Timed out waiting for delete event to be processed")
	}
}
