# k8s-ssh-router

`k8s-ssh-router` is a Go application designed to handle SSH connections and forward them to specific services within a Kubernetes cluster. It uses Kubernetes secrets for authentication and supports various SSH functionalities, including SFTP.

## Features

- **SSH Authentication:** Uses Kubernetes secrets for user authentication.
- **Forwarding:** Forwards SSH connections to specific services in the cluster.
- **SFTP Support:** Supports file transfers via SFTP.
- **Metrics:** Exposes Prometheus metrics for active sessions.
- **Configurable:** Various options can be configured via command-line arguments or environment variables.

## Why This Solution?

### Secure SSH Access

SSH does not natively support TLS, making it less secure compared to modern protocols that do. This solution ensures secure access by handling authentication and authorization within the Kubernetes cluster, leveraging Kubernetes secrets for storing user credentials.

### Cost-Effective Scaling

Running multiple LoadBalancer services can be expensive and may not scale well due to IP address limitations. This solution uses a single LoadBalancer service to route SSH traffic to multiple backend pods, significantly reducing costs and simplifying management.

### Auto-Scaling

Managing a fixed number of SSH servers can lead to either over-provisioning or under-provisioning of resources. This solution uses Prometheus metrics to dynamically scale the number of SSH gateway pods based on actual usage, ensuring optimal resource utilization.

## Table of Contents

- [k8s-ssh-router](#k8s-ssh-router)
  - [Features](#features)
  - [Why This Solution?](#why-this-solution)
    - [Secure SSH Access](#secure-ssh-access)
    - [Cost-Effective Scaling](#cost-effective-scaling)
    - [Auto-Scaling](#auto-scaling)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Running the application](#running-the-application)
    - [Configuration](#configuration)
  - [Development](#development)
    - [Prerequisites](#prerequisites)
      - [Running Tests](#running-tests)
      - [Building the Docker Image](#building-the-docker-image)
  - [CI/CD Pipeline](#cicd-pipeline)
  - [Dependencies Management](#dependencies-management)
  - [Contributing](#contributing)
  - [License](#license)

## Installation

To install the `k8s-ssh-router`, you need to have Go installed. You can then build the application from source.

```bash
git clone https://github.com/davidcollom/k8s-ssh-router.git
cd k8s-ssh-router
go build -o k8s-ssh-router ./cmd
```

You can also pull the Docker image from GitHub Container Registry (GHCR):

```sh
docker pull ghcr.io/davidcollom/k8s-ssh-router:latest
```


## Usage

### Running the application

To run the application, you can use the built binary:

```sh
./k8s-ssh-router --reconcile-interval 60 --ssh-port 2222 --metrics-port 9090 --namespace default --private-key-path /path/to/id_rsa
```

Or you can run it using Docker:

```sh
docker run -d -p 2222:2222 -p 9090:9090 \
  -e RECONCILE_INTERVAL=60 \
  -e SSH_PORT=2222 \
  -e METRICS_PORT=9090 \
  -e NAMESPACE=default \
  -e PRIVATE_KEY_PATH=/path/to/id_rsa \
  ghcr.io/davidcollom/k8s-ssh-router:latest
```

### Configuration

The following options can be configured via command-line arguments or environment variables:

- `--reconcile-interval` / `RECONCILE_INTERVAL`: Reconciliation interval in seconds (default: 60)
- `--ssh-port` / `SSH_PORT`: SSH server port (default: 2222)
- `--metrics-port` / `METRICS_PORT`: Metrics server port (default: 9090)
- `--namespace` / `NAMESPACE`: Kubernetes namespace
- `--private-key-path` / `PRIVATE_KEY_PATH`: Path to the private key file

## Development

### Prerequisites

- Go 1.21 or later
- Docker

#### Running Tests

To run the tests locally:

```sh
go tst ./... -v
```

#### Building the Docker Image

To build the Docker image:

```sh
docker build -t ghcr.io/davidcollom/k8s-ssh-router:latest .
```


## CI/CD Pipeline

This project uses GitHub Actions for continuous integration and deployment. The workflow is defined in `.github/workflows/go.yml`.

## Dependencies Management

This project uses Dependabot to keep dependencies up to date. The configuration is defined in `.github/dependabot.yml`.


## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any changes.

## License

This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for details.
