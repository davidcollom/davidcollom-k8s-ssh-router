package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Config struct {
	LabelSelector string `envconfig:"LABEL_SELECTOR" default:"ssh=users"`
	Namespace     string `envconfig:"NAMESPACE" default:"default"`
}

type SecretData struct {
	Password string
	Key      string
	Service  string
}

func loadKubernetesSecrets(config *Config) (map[string]SecretData, error) {
	configK8s, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(configK8s)
	if err != nil {
		return nil, err
	}

	secrets, err := clientset.CoreV1().Secrets(config.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: config.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	secretDataMap := make(map[string]SecretData)
	for _, secret := range secrets.Items {
		username := secret.Labels["username"]
		password := string(secret.Data["pass"])
		key := string(secret.Data["key"])
		service := string(secret.Data["service"])

		secretDataMap[username] = SecretData{
			Password: password,
			Key:      key,
			Service:  service,
		}
	}

	return secretDataMap, nil
}

func authenticate(username, password string, secretDataMap map[string]SecretData) (SecretData, bool) {
	secretData, exists := secretDataMap[username]
	if !exists {
		return SecretData{}, false
	}

	if secretData.Password == password {
		return secretData, true
	}

	return SecretData{}, false
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err)
	}

	username := os.Getenv("PAM_USER")
	password := os.Getenv("PAM_AUTHTOK")

	secretDataMap, err := loadKubernetesSecrets(&config)
	if err != nil {
		log.Fatal(err)
	}

	_, authenticated := authenticate(username, password, secretDataMap)
	if !authenticated {
		log.Fatalf("Authentication failed for user %s", username)
	}

	fmt.Printf("Authentication successful for user %s\n", username)
}
