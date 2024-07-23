import pytest
from unittest.mock import patch, MagicMock
import sys
import os

# Add the scripts directory to the sys.path
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../../scripts')))

import pam_k8s_auth

@pytest.fixture
def mock_pamh():
    pamh = MagicMock()
    pamh.get_user.return_value = 'test-user'
    pamh.authtok = 'test-pass'
    return pamh

@pytest.fixture
def user_info():
    return {
        'password': 'test-pass',
        'service': 'test-service.test-namespace.svc.cluster.local'
    }

@patch('pam_k8s_auth.cache')
def test_pam_sm_authenticate_success(mock_cache, mock_pamh, user_info):
    mock_cache.get.return_value = user_info

    result = pam_k8s_auth.pam_sm_authenticate(mock_pamh, None, None)
    assert result == pam_k8s_auth.pamh.PAM_SUCCESS
    assert mock_pamh.env['K8S_SERVICE'] == 'test-service.test-namespace.svc.cluster.local'

@patch('pam_k8s_auth.cache')
def test_pam_sm_authenticate_failure(mock_cache, mock_pamh, user_info):
    mock_cache.get.return_value = user_info
    mock_pamh.authtok = 'wrong-pass'

    result = pam_k8s_auth.pam_sm_authenticate(mock_pamh, None, None)
    assert result == pam_k8s_auth.pamh.PAM_AUTH_ERR
