package k8s

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Executor interface {
	Stream(options remotecommand.StreamOptions) error
}

func ExecInPod(clientset kubernetes.Interface, restClient rest.Interface, executor Executor, config *rest.Config, username, command string, conn ssh.Channel, isTerminal bool) error {
	fmt.Printf("Cache: %v \n", localCache.ItemCount())
	userSecret, found := GetSecretFromCache(username)
	if !found {
		return fmt.Errorf("user secret not found in cache")
	}
	secret := userSecret.(map[string]string)

	service := secret["service"]
	podLabelSelector := secret["podLabelSelector"]
	containerName := secret["containerName"]
	shell := secret["shell"]
	if shell == "" {
		shell = "/bin/sh"
	}

	pods, err := clientset.CoreV1().Pods(service).List(context.TODO(), metav1.ListOptions{
		LabelSelector: podLabelSelector,
	})
	if err != nil || len(pods.Items) == 0 {
		return fmt.Errorf("failed to list pods: %v", err)
	}
	pod := pods.Items[0]

	req := restClient.
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		Param("container", containerName).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", strconv.FormatBool(isTerminal))

	if command != "" {
		req.Param("command", shell).Param("command", "-c").Param("command", command)
	} else {
		req.Param("command", shell)
	}

	return executor.Stream(remotecommand.StreamOptions{
		Stdin:  conn,
		Stdout: conn,
		Stderr: conn.Stderr(),
		Tty:    isTerminal,
	})
}
