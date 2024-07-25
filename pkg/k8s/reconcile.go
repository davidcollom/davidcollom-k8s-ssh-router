package k8s

import (
	"context"
	"fmt"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func startReconciliation(clientset kubernetes.Interface, reconcileInterval int, namespace string) {
	ticker := time.NewTicker(time.Duration(reconcileInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		reconcileCache(clientset, namespace)
	}
}

func reconcileCache(clientset kubernetes.Interface, namespace string) {
	log.Println("Starting reconciliation...")
	secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "ssh=user",
	})
	if err != nil {
		log.Printf("failed to list secrets: %v\n", err)
		return
	}

	currentSecrets := make(map[string]bool)
	for _, secret := range secrets.Items {
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
		SetSecretInCache(usernameWithNamespace, secretData)
		currentSecrets[usernameWithNamespace] = true
	}

	// Remove any secrets from the cache that are no longer present in the cluster
	for key := range localCache.Items() {
		if !currentSecrets[key] {
			DeleteSecretFromCache(key)
		}
	}

	log.Println("Reconciliation completed")
}
