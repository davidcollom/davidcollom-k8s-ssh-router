# Kubernetes SSH Gateway with Transparent Authentication and Autoscaling

## Overview

This project provides a scalable, cost-effective solution for securely accessing Kubernetes pods via SSH. By leveraging a central SSH gateway, Kubernetes secrets, and Prometheus metrics, this solution enables transparent authentication and automatic scaling of SSH sessions to pods within your cluster.

## Features

- **Transparent Authentication**: User credentials are stored in Kubernetes secrets and are transparently authenticated without any special configuration on the client's machine.
- **Cost-Effective Scaling**: Uses a single LoadBalancer service to route SSH traffic, reducing costs and simplifying IP address management.
- **Auto-Scaling**: Automatically scales the number of SSH gateway pods based on the number of active SSH sessions, ensuring efficient resource utilization.
- **Prometheus Monitoring**: Exposes custom metrics for monitoring and autoscaling using Prometheus and the Prometheus Adapter.

## Why This Solution?

### Secure SSH Access

SSH does not natively support TLS, making it less secure compared to modern protocols that do. This solution ensures secure access by handling authentication and authorization within the Kubernetes cluster, leveraging Kubernetes secrets for storing user credentials.

### Cost-Effective Scaling

Running multiple LoadBalancer services can be expensive and may not scale well due to IP address limitations. This solution uses a single LoadBalancer service to route SSH traffic to multiple backend pods, significantly reducing costs and simplifying management.

### Auto-Scaling

Managing a fixed number of SSH servers can lead to either over-provisioning or under-provisioning of resources. This solution uses Prometheus metrics to dynamically scale the number of SSH gateway pods based on actual usage, ensuring optimal resource utilization.

## How It Works

### Architecture

1. **SSH Gateway Pod**: Users connect to an SSH gateway pod via a LoadBalancer service. The SSH gateway pod is configured to authenticate users against credentials stored in Kubernetes secrets.
2. **Transparent Authentication**: A custom PAM module (`pam_k8s_auth.py`) fetches user credentials from Kubernetes secrets and sets the appropriate environment variables for routing the SSH connection.
3. **SSH Forwarding**: An SSH forwarding script (`ssh_forward.sh`) uses the environment variables to route the authenticated SSH session to the appropriate pod within the cluster.
4. **Prometheus Monitoring**: A Prometheus exporter (`ssh_prometheus_exporter.py`) exposes custom metrics, such as the number of active SSH sessions, which are used for monitoring and autoscaling.
5. **Autoscaling**: The Horizontal Pod Autoscaler (HPA) uses custom metrics exposed by the Prometheus Adapter to scale the number of SSH gateway pods based on the number of active SSH sessions.

### Components

- **Kubernetes Secrets**: Store user credentials including username, password, and SSH keys.
- **Custom PAM Module**: Authenticates users against Kubernetes secrets.
- **SSH Forwarding Script**: Routes authenticated SSH sessions to the appropriate pod.
- **Prometheus Exporter**: Exposes custom metrics for monitoring and autoscaling.
- **Prometheus Adapter**: Makes Prometheus metrics available to the Kubernetes custom metrics API.
- **Horizontal Pod Autoscaler (HPA)**: Automatically scales the SSH gateway pods based on custom metrics.

## Setup and Deployment

### Prerequisites

- Kubernetes cluster
- Prometheus and Prometheus Adapter installed
- Docker and Docker Hub or GitHub Container Registry (GHCR) for building and storing Docker images

### Steps

1. **Build and Push Docker Image**

   ```sh
   docker build -t ghcr.io/<your-username>/ssh-server:latest .
   docker push ghcr.io/<your-username>/ssh-server:latest
   ```

2. **Install Prometheus Adapter via Helm**
  Add the Prometheus community Helm repository:

  ```sh
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
  ```

  Install the Prometheus Adapter:

  ```sh
  helm install prometheus-adapter prometheus-community/prometheus-adapter --namespace custom-metrics --create-namespace
  ```

  Customize the Prometheus Adapter configuration if needed by creating a values.yaml file:
  values.yaml

  ```yaml
  prometheus:
  url: http://prometheus-server.prometheus.svc.cluster.local

rules:
  default: false

  custom:
    - seriesQuery: 'ssh_active_sessions'
      resources:
        overrides:
          namespace: {resource: "namespace"}
      name:
        matches: "^(.*)_total"
        as: "${1}_per_second"
      metricsQuery: 'sum(rate(ssh_active_sessions[2m])) by (namespace)'
  ```

  Install with custom values:

  ```sh
  helm install prometheus-adapter prometheus-community/prometheus-adapter --namespace custom-metrics --create-namespace -f values.yaml
  ```

3. **Apply Kubernetes Manifests**
  Apply the SSH service and deployment manifests:

  ```sh
  kubectl apply -f deploy/
  ```
