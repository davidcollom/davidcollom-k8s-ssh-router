package sshserver

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/davidcollom/k8s-ssh-router/pkg/auth"
	"github.com/davidcollom/k8s-ssh-router/pkg/k8s"
	"github.com/davidcollom/k8s-ssh-router/pkg/metrics"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func RunServer(reconcileInterval, sshPort, metricsPort int, namespace, privateKeyPath string) {
	go func() {
		readyCh := make(chan struct{})
		if _, err := k8s.WatchSecretsClusterWide(reconcileInterval, namespace, readyCh); err != nil {
			log.Fatalf("Failed to start watcher: %v", err)
		}
	}()
	go metrics.StartMetricsServer(metricsPort)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create in-cluster config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	startSSHServer(sshPort, privateKeyPath, clientset, config)
}

func startSSHServer(port int, privateKeyPath string, clientset kubernetes.Interface, restConfig *rest.Config) {
	sshConfig := &ssh.ServerConfig{
		NoClientAuth:      false,
		PasswordCallback:  auth.PasswordCallback,
		PublicKeyCallback: auth.PublicKeyCallback,
	}

	privateBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to load private key from %s: %v", privateKeyPath, err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	sshConfig.AddHostKey(private)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}
	log.Printf("Listening on 0.0.0.0:%d...", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}
		go HandleSSHConnection(conn, sshConfig, clientset, restConfig)
	}
}
