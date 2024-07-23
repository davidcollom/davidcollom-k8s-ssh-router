import pytest
from unittest.mock import patch, mock_open
import sys
import os

# Add the scripts directory to the sys.path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../scripts')))

import ssh_prometheus_exporter

@pytest.fixture
def mock_file_content():
    return 'sshd: Accepted session\nsshd: session closed\n'

@patch('builtins.open', new_callable=mock_open, read_data='sshd: Accepted session\nsshd: session closed')
def test_count_active_sessions(mock_file, mock_file_content):
    result = ssh_prometheus_exporter.count_active_sessions('/var/log/auth.log')
    assert result == 0

@patch('ssh_prometheus_exporter.count_active_sessions', return_value=5)
@patch('ssh_prometheus_exporter.ACTIVE_SESSIONS')
def test_update_metrics(mock_active_sessions, mock_count_active_sessions):
    ssh_prometheus_exporter.update_metrics('/var/log/auth.log')
    mock_active_sessions.set.assert_called_once_with(5)
