package sshserver

import (
	"io"
	"log"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// handleSFTP handles SFTP requests.
func handleSFTP(channel ssh.Channel, clientset kubernetes.Interface, config *rest.Config, namespace, podName, containerName string) {
	server, err := sftp.NewServer(channel)
	if err != nil {
		log.Printf("Failed to create SFTP server: %v", err)
		channel.Close()
		return
	}

	handler := &SFTPHandler{
		Clientset:     clientset,
		Config:        config,
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
	}

	server.SetFileGet(handler)
	server.SetFilePut(handler)
	server.SetFileRemove(handler)
	server.SetFileList(handler)
	server.SetFileRename(handler)
	server.SetFileChmod(handler)

	if err := server.Serve(); err == io.EOF {
		server.Close()
		log.Printf("SFTP client exited session.")
	} else if err != nil {
		log.Printf("SFTP server completed with error: %v", err)
		channel.Close()
	}
}
