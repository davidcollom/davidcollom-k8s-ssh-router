import base64
import time
from kubernetes import client, config
from diskcache import Cache

# Initialize the cache
cache = Cache('/var/tmp/k8s_ssh_cache')

def fetch_k8s_secrets():
    config.load_incluster_config()
    v1 = client.CoreV1Api()
    secrets = v1.list_secret_for_all_namespaces(label_selector='ssh=users')
    return secrets.items

def update_cache():
    while True:
        try:
            secrets = fetch_k8s_secrets()
            with cache.transact():
                cache.clear()
                for secret in secrets:
                    namespace = secret.metadata.namespace
                    name = secret.metadata.name
                    override_user = secret.data.get('user')
                    user = f"{namespace}-{name}" if not override_user else base64.b64decode(override_user).decode('utf-8')
                    passw = base64.b64decode(secret.data.get('pass')).decode('utf-8')
                    key = base64.b64decode(secret.data.get('key')).decode('utf-8')
                    service = base64.b64decode(secret.data.get('service')).decode('utf-8')
                    service_with_namespace = f"{service}.{namespace}.svc.cluster.local"
                    cache[user] = {'password': passw, 'key': key, 'service': service_with_namespace}
            print("Cache updated successfully.")
        except Exception as e:
            print(f"Error updating cache: {e}")
        time.sleep(300)  # Update cache every 5 minutes

if __name__ == "__main__":
    update_cache()
