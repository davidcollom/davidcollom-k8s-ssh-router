#!/bin/bash

# Get the target service from the environment variable
TARGET_SERVICE=$K8S_SERVICE

if [ -z "$TARGET_SERVICE" ]; then
  echo "No target service specified."
  exit 1
fi

# Use SSH to forward the connection to the internal service
exec ssh -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null -l "$PAM_USER" "$TARGET_SERVICE"
