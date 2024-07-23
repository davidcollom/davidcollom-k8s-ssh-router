FROM ubuntu:20.04

# Set DEBIAN_FRONTEND to noninteractive to avoid dialog frontend issues
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y openssh-server python3-pip libpam-python

# Install necessary Python libraries
RUN pip3 install kubernetes diskcache prometheus_client

# Create necessary directories
RUN mkdir /var/run/sshd

# Create a non-root user
RUN useradd -m -s /bin/bash sshuser

# Set up SSH server configuration for non-root user
RUN mkdir /home/sshuser/.ssh
RUN chown -R sshuser:sshuser /home/sshuser/.ssh
RUN echo 'Port 2222' >> /etc/ssh/sshd_config
RUN echo 'PermitRootLogin no' >> /etc/ssh/sshd_config
RUN echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config
RUN echo 'UsePAM yes' >> /etc/ssh/sshd_config
RUN echo 'AcceptEnv K8S_SERVICE' >> /etc/ssh/sshd_config

# Copy scripts and configuration with ownership set to sshuser
COPY --chown=sshuser:sshuser scripts/pam_k8s_auth.py /usr/local/bin/pam_k8s_auth.py
COPY --chown=sshuser:sshuser scripts/cache_updater.py /usr/local/bin/cache_updater.py
COPY --chown=sshuser:sshuser scripts/ssh_forward.sh /usr/local/bin/ssh_forward.sh
COPY --chown=sshuser:sshuser scripts/ssh_prometheus_exporter.py /usr/local/bin/ssh_prometheus_exporter.py

# Ensure scripts are executable
RUN chmod +x /usr/local/bin/pam_k8s_auth.py \
             /usr/local/bin/cache_updater.py \
             /usr/local/bin/ssh_forward.sh \
             /usr/local/bin/ssh_prometheus_exporter.py

# Copy PAM configuration
RUN echo '#%PAM-1.0\nauth requisite pam_python.so /usr/local/bin/pam_k8s_auth.py\nsession optional pam_exec.so /usr/local/bin/ssh_forward.sh' > /etc/pam.d/sshd

EXPOSE 2222 8000

# Switch to the non-root user
USER sshuser

# Default entrypoint
CMD ["/usr/sbin/sshd", "-D"]
