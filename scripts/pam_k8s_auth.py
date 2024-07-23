import base64
import os
from diskcache import Cache

# Initialize the cache
cache = Cache('/var/tmp/k8s_ssh_cache')

def pam_sm_authenticate(pamh, flags, argv):
    username = pamh.get_user(None)
    password = pamh.authtok

    if not username or not password:
        return pamh.PAM_AUTH_ERR

    user_info = cache.get(username)
    if not user_info:
        return pamh.PAM_AUTH_ERR

    if password and user_info.get('password') == password:
        pamh.env['K8S_SERVICE'] = user_info.get('service')
        return pamh.PAM_SUCCESS

    return pamh.PAM_AUTH_ERR

def pam_sm_setcred(pamh, flags, argv):
    return pamh.PAM_SUCCESS
