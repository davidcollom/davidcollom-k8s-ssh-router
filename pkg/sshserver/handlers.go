package sshserver

import (
	"log"
	"net"

	"github.com/davidcollom/k8s-ssh-router/pkg/k8s"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func handleSSHRequests(clientset kubernetes.Interface, restClient rest.Interface, config *rest.Config, executor k8s.Executor, channel ssh.Channel, requests <-chan *ssh.Request, username string) {
	isTerminal := false
	for req := range requests {
		switch req.Type {
		case "pty-req":
			isTerminal = true
			req.Reply(true, nil)
		case "exec":
			command := string(req.Payload[4:])
			log.Printf("Received exec request: %s", command)
			if err := k8s.ExecInPod(clientset, restClient, executor, config, username, command, channel, isTerminal); err != nil {
				log.Printf("Exec in pod failed: %v", err)
				channel.Stderr().Write([]byte(err.Error()))
			}
			channel.Close()
		case "shell":
			log.Printf("Received shell request")
			if err := k8s.ExecInPod(clientset, restClient, executor, config, username, "", channel, isTerminal); err != nil {
				log.Printf("Exec in pod failed: %v", err)
				channel.Stderr().Write([]byte(err.Error()))
			}
			channel.Close()
		// case "subsystem":
		// 	subsystem := string(req.Payload[4:])
		// 	if subsystem == "sftp" {
		// 		log.Printf("Received sftp request")
		// 		req.Reply(true, nil)
		// 		go handleSFTP(channel)
		// 	} else {
		// 		log.Printf("Unknown subsystem request: %s", subsystem)
		// 		req.Reply(false, nil)
		// 	}
		default:
			log.Printf("Unknown request type: %s", req.Type)
			req.Reply(false, nil)
		}
	}
}

func HandleSSHConnection(conn net.Conn, sshConfig *ssh.ServerConfig, clientset kubernetes.Interface, restConfig *rest.Config) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		log.Printf("Failed to create REST client: %v", err)
		return
	}

	for newChannel := range chans {
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		go handleSSHRequests(clientset, restClient, restConfig, nil, channel, requests, sshConn.User())
	}
}
