FROM golang:1.22 AS builder

WORKDIR /app
COPY . /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pam_k8s_auth .

FROM ubuntu:20.04

# Set DEBIAN_FRONTEND to noninteractive to avoid dialog frontend issues
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y openssh-server libpam0g-dev

COPY --from=builder /app/pam_k8s_auth /usr/local/bin/pam_k8s_auth

# Configure SSH server
RUN echo 'Port 2222' >> /etc/ssh/sshd_config
RUN echo 'PermitRootLogin no' >> /etc/ssh/sshd_config
RUN echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config
RUN echo 'UsePAM yes' >> /etc/ssh/sshd_config
RUN echo 'AcceptEnv K8S_SERVICE' >> /etc/ssh/sshd_config

# Copy PAM configuration
RUN echo '#%PAM-1.0\nauth requisite pam_exec.so /usr/local/bin/pam_k8s_auth' > /etc/pam.d/sshd

EXPOSE 2222

CMD ["/usr/sbin/sshd", "-D"]
