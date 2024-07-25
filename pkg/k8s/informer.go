package k8s

import (
	"context"
	"fmt"
	"log"

	"github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getClientset() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			log.Fatalf("failed to create Kubernetes client config: %v", err)
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

func WatchSecretsClusterWide(reconcileInterval int, namespace string, readyCh chan struct{}) (watch.Interface, error) {
	clientset, err := getClientset()
	if err != nil {
		return nil, err
	}

	go startReconciliation(clientset, reconcileInterval, namespace)

	watcher, err := clientset.CoreV1().Secrets(namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: "ssh=user",
	})
	if err != nil {
		log.Fatalf("failed to set up secret watcher: %v", err)
		return nil, err
	}

	log.Println("Starting secret watcher...")
	readyCh <- struct{}{}
	ch := watcher.ResultChan()
	go func() {
		for event := range ch {
			switch event.Type {
			case watch.Added:
				log.Println("Processing added event")
				secret := event.Object.(*corev1.Secret)
				processSecret(secret)
			case watch.Modified:
				log.Println("Processing modified event")
				secret := event.Object.(*corev1.Secret)
				processSecret(secret)
			case watch.Deleted:
				log.Println("Processing deleted event")
				secret := event.Object.(*corev1.Secret)
				namespace := secret.Namespace
				username := string(secret.Data["username"])
				usernameWithNamespace := fmt.Sprintf("%s-%s", namespace, username)
				localCache.Delete(usernameWithNamespace)
				log.Printf("Deleted secret: %s, cache size: %d\n", usernameWithNamespace, localCache.ItemCount())
			}
		}
	}()
	return watcher, nil
}

func processSecret(secret *corev1.Secret) {
	namespace := secret.Namespace
	username := string(secret.Data["username"])
	usernameWithNamespace := fmt.Sprintf("%s-%s", namespace, username)
	secretData := map[string]string{
		"password":         string(secret.Data["password"]),
		"publicKey":        string(secret.Data["publicKey"]),
		"service":          string(secret.Data["service"]),
		"podLabelSelector": string(secret.Data["podLabelSelector"]),
		"containerName":    string(secret.Data["containerName"]),
		"shell":            string(secret.Data["shell"]),
	}
	localCache.Set(usernameWithNamespace, secretData, cache.DefaultExpiration)
	log.Printf("Added/Modified secret: %s, cache size: %d\n", usernameWithNamespace, localCache.ItemCount())
}
