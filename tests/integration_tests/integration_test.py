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

    # Deploy the client containers in separate namespaces
    deploy_client_containers(image_tag)

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
    teardown_client_containers()

def deploy_client_containers(image_tag):
    namespaces = ["client-ns-1", "client-ns-2"]
    for ns in namespaces:
        subprocess.run(["kubectl", "create", "namespace", ns], check=True)

    client_1_yaml = f"""
apiVersion: v1
kind: Pod
metadata:
  name: client-1
  namespace: client-ns-1
  labels:
    app: ssh-client
spec:
  initContainers:
  - name: init-client
    image: alpine
    command: ["/bin/sh", "-c", "echo 'Client 1' > /mnt/echo1.txt"]
    volumeMounts:
    - mountPath: /mnt
      name: client-mount
  containers:
  - name: ssh-client
    image: ghcr.io/{os.getenv('GITHUB_REPOSITORY')}/ssh-client:{image_tag}
    volumeMounts:
    - mountPath: /mnt
      name: client-mount
  volumes:
  - name: client-mount
    emptyDir: {{}}
"""

    client_2_yaml = f"""
apiVersion: v1
kind: Pod
metadata:
  name: client-2
  namespace: client-ns-2
  labels:
    app: ssh-client
spec:
  initContainers:
  - name: init-client
    image: alpine
    command: ["/bin/sh", "-c", "echo 'Client 2' > /mnt/echo2.txt"]
    volumeMounts:
    - mountPath: /mnt
      name: client-mount
  containers:
  - name: ssh-client
    image: ghcr.io/{os.getenv('GITHUB_REPOSITORY')}/ssh-client:{image_tag}
    volumeMounts:
    - mountPath: /mnt
      name: client-mount
  volumes:
  - name: client-mount
    emptyDir: {{}}
"""
    with open("client_1.yaml", "w") as f:
        f.write(client_1_yaml)

    with open("client_2.yaml", "w") as f:
        f.write(client_2_yaml)

    subprocess.run(["kubectl", "apply", "-f", "client_1.yaml"], check=True)
    subprocess.run(["kubectl", "apply", "-f", "client_2.yaml"], check=True)

def teardown_client_containers():
    subprocess.run(["kubectl", "delete", "namespace", "client-ns-1"], check=True)
    subprocess.run(["kubectl", "delete", "namespace", "client-ns-2"], check=True)
    os.remove("client_1.yaml")
    os.remove("client_2.yaml")

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
                            capture_output=True, text​⬤
