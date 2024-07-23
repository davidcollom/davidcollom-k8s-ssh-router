import re
import time
from prometheus_client import start_http_server, Gauge

# Define Prometheus metrics
ACTIVE_SESSIONS = Gauge('ssh_active_sessions', 'Number of active SSH sessions')

def count_active_sessions(log_file_path):
    # Regex patterns to match session open and close logs
    session_open_pattern = re.compile(r'sshd.*Accepted.*session')
    session_close_pattern = re.compile(r'sshd.*session closed')

    active_sessions = 0

    with open(log_file_path, 'r') as log_file:
        for line in log_file:
            if session_open_pattern.search(line):
                active_sessions += 1
            elif session_close_pattern.search(line):
                active_sessions -= 1

    return max(active_sessions, 0)

def update_metrics(log_file_path):
    active_sessions = count_active_sessions(log_file_path)
    ACTIVE_SESSIONS.set(active_sessions)

if __name__ == '__main__':
    log_file_path = '/var/log/auth.log'  # Update this path as per your system
    start_http_server(8000)  # Expose metrics on port 8000

    while True:
        update_metrics(log_file_path)
        time.sleep(10)  # Update metrics every 10 seconds
