import base64
import pytest
from unittest.mock import patch, MagicMock
import sys
import os

# Add the scripts directory to the sys.path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../scripts')))

import cache_updater

@pytest.fixture
def mock_secret():
    secret = MagicMock()
    secret.metadata.namespace = 'test-namespace'
    secret.metadata.name = 'test-name'
    secret.data = {
        'user': base64.b64encode(b'test-user').decode('utf-8'),
        'pass': base64.b64encode(b'test-pass').decode('utf-8'),
        'key': base64.b64encode(b'test-key').decode('utf-8'),
        'service': base64.b64encode(b'test-service').decode('utf-8')
    }
    return secret

@patch('cache_updater.client.CoreV1Api')
@patch('cache_updater.config.load_incluster_config')
def test_fetch_k8s_secrets(mock_load_config, mock_v1_api, mock_secret):
    mock_v1_api.return_value.list_secret_for_all_namespaces.return_value.items = [mock_secret]

    secrets = cache_updater.fetch_k8s_secrets()
    assert len(secrets) == 1
    assert secrets[0].metadata.namespace == 'test-namespace'
    assert secrets[0].metadata.name == 'test-name'

@patch('cache_updater.fetch_k8s_secrets')
@patch('cache_updater.cache')
def test_update_cache(mock_cache, mock_fetch_secrets, mock_secret):
    mock_fetch_secrets.return_value = [mock_secret]

    cache_updater.update_cache()

    assert mock_cache.__getitem__.called
    assert mock_cache.__setitem__.called
    assert mock_cache.__setitem__.call_args[0][0] == 'test-namespace-test-name'
    assert mock_cache.__setitem__.call_args[0][1]['password'] == 'test-pass'
    assert mock_cache.__setitem__.call_args[0][1]['service'] == 'test-service.test-namespace.svc.cluster.local'
