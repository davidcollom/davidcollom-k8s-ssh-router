package sshserver

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type SFTPHandler struct {
	Clientset     kubernetes.Interface
	Config        *rest.Config
	Namespace     string
	PodName       string
	ContainerName string
}

func (h *SFTPHandler) FileGet(r *sftp.Request) (io.ReaderAt, error) {
	cmd := fmt.Sprintf("cat %s", r.Filepath)
	stdout, stderr, err := h.execInPod(cmd)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v, stderr: %s", err, stderr)
	}
	return bytes.NewReader(stdout), nil
}

func (h *SFTPHandler) FilePut(r *sftp.Request) (io.WriterAt, error) {
	cmd := fmt.Sprintf("cat > %s", r.Filepath)
	stdin := new(bytes.Buffer)
	writer := &sftpFileWriter{stdin: stdin, handler: h, cmd: cmd}
	return writer, nil
}

func (h *SFTPHandler) FileRemove(r *sftp.Request) error {
	cmd := fmt.Sprintf("rm %s", r.Filepath)
	_, stderr, err := h.execInPod(cmd)
	if err != nil {
		return fmt.Errorf("error removing file: %v, stderr: %s", err, stderr)
	}
	return nil
}

func (h *SFTPHandler) FileList(r *sftp.Request) ([]os.FileInfo, error) {
	cmd := fmt.Sprintf("ls -l %s", r.Filepath)
	stdout, stderr, err := h.execInPod(cmd)
	if err != nil {
		return nil, fmt.Errorf("error listing directory: %v, stderr: %s", err, stderr)
	}
	// Parse stdout to create os.FileInfo slice
	// ...
	return nil, nil
}

func (h *SFTPHandler) FileRename(r *sftp.Request) error {
	cmd := fmt.Sprintf("mv %s %s", r.Filepath, r.Target)
	_, stderr, err := h.execInPod(cmd)
	if err != nil {
		return fmt.Errorf("error renaming file: %v, stderr: %s", err, stderr)
	}
	return nil
}

func (h *SFTPHandler) FileChmod(r *sftp.Request) error {
	cmd := fmt.Sprintf("chmod %o %s", r.Attributes().Permissions, r.Filepath)
	_, stderr, err := h.execInPod(cmd)
	if err != nil {
		return fmt.Errorf("error changing file permissions: %v, stderr: %s", err, stderr)
	}
	return nil
}

func (h *SFTPHandler) execInPod(cmd string) ([]byte, []byte, error) {
	req := h.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(h.PodName).
		Namespace(h.Namespace).
		SubResource("exec").
		Param("container", h.ContainerName).
		Param("command", "/bin/sh").
		Param("command", "-c").
		Param("command", cmd).
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false")

	exec, err := remotecommand.NewSPDYExecutor(h.Config, "POST", req.URL())
	if err != nil {
		return nil, nil, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		return nil, nil, err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

type sftpFileWriter struct {
	stdin   *bytes.Buffer
	handler *SFTPHandler
	cmd     string
}

func (w *sftpFileWriter) WriteAt(p []byte, off int64) (int, error) {
	_, err := w.stdin.Write(p)
	return len(p), err
}

func (w *sftpFileWriter) Close() error {
	_, stderr, err := w.handler.execInPod(w.cmd)
	if err != nil {
		return fmt.Errorf("error writing file: %v, stderr: %s", err, stderr)
	}
	return nil
}
