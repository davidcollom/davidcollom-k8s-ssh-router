import pytest
import subprocess
import time
import os

@pytest.fixture(scope="module")
def setup_kubernetes():
    image_tag = os.getenv("IMAGE_TAG")

    # Apply Kubernetes manifests from the examples directory
    subprocess.run(["kubectl", "apply", "-f", "examples/user-secret.yaml"], check=True)
    subprocess.run(["kubectl", "apply", "-f", "examples/pvc.yaml"], check=True)
    subprocess.run(["kubectl", "apply", "-f", "examples/service.yaml"], check=True)

    # Apply the StatefulSet manifest and patch the image tag
    subprocess.run(["kubectl", "apply", "-f", "examples/statefulset.yaml"], check=True)
    subprocess.run(["kubectl", "set", "image", "statefulset/ssh-statefulset", f"ssh-server=ghcr.io/{os.getenv('GITHUB_REPOSITORY')}/ssh-server:{image_tag}"], check=True)

    # Apply the frontend services/router manifests from the deploy directory
    subprocess.run(["kubectl", "apply", "-f", "deploy/deployment.yaml"], check=True)
    subprocess.run(["kubectl", "apply", "-f", "deploy/service.yaml"], check=True)
    subprocess.run(["kubectl", "apply", "-f", "deploy/hpa.yaml"], check=True)
    subprocess.run(["kubectl", "apply", "-f", "deploy/prometheus-adapter-config.yaml"], check=True)

    # Wait for the StatefulSet and Deployment to be ready
    wait_for_statefulset("ssh-statefulset")
    wait_for_deployment("ssh-service")

    yield

    # Teardown Kubernetes resources
    subprocess.run(["kubectl", "delete", "-f", "examples/user-secret.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "examples/pvc.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "examples/service.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "examples/statefulset.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "deploy/deployment.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "deploy/service.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "deploy/hpa.yaml"], check=True)
    subprocess.run(["kubectl", "delete", "-f", "deploy/prometheus-adapter-config.yaml"], check=True)

def wait_for_statefulset(statefulset_name, namespace="default", timeout=600):
    start_time = time.time()
    while time.time() - start_time < timeout:
        status = subprocess.run(["kubectl", "rollout", "status", f"statefulset/{statefulset_name}", "-n", namespace],
                                capture_output=True, text=True)
        if "successfully rolled out" in status.stdout:
            return True
        time.sleep(10)
    raise TimeoutError(f"StatefulSet {statefulset_name} not ready after {timeout} seconds")

def wait_for_deployment(deployment_name, namespace="default", timeout=600):
    start_time = time.time()
    while time.time() - start_time < timeout:
        status = subprocess.run(["kubectl", "rollout", "status", f"deployment/{deployment_name}", "-n", namespace],
                                capture_output=True, text=True)
        if "successfully rolled out" in status.stdout:
            return True
        time.sleep(10)
    raise TimeoutError(f"Deployment {deployment_name} not ready after {timeout} seconds")

def test_pods_running(setup_kubernetes):
    result = subprocess.run(["kubectl", "get", "pods", "-l", "app=ssh-statefulset", "-o", "jsonpath='{.items[*].status.phase}'"],
                            capture_output=True, text=True)
    assert "Running" in result.stdout

def test_service_endpoint(setup_kubernetes):
    result = subprocess.run(["kubectl", "get", "svc", "ssh-statefulset-service", "-o", "jsonpath='{.status.loadBalancer.ingress[0].ip}'"],
                            capture_output=True, text=True)
    service_ip = result.stdout.strip("'")
    assert service_ip != ""

    # Perform an SSH connection test (assuming the SSH server responds to a simple SSH command)
    ssh_test = subprocess.run(["sshpass", "-p", "example-password", "ssh", "-oStrictHostKeyChecking=no", f"example-user@{service_ip}", "echo", "SSH Connection Successful"],
                              capture_output=True, text=True)
    assert "SSH Connection Successful" in ssh_test.stdout
