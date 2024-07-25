package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/davidcollom/k8s-ssh-router/pkg/k8s"

	"golang.org/x/crypto/ssh"
)

func PasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	err := authenticateUser(username, string(password), "")
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func PublicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	username := conn.User()
	clientKey := string(ssh.MarshalAuthorizedKey(key))
	err := authenticateUser(username, "", clientKey)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func authenticateUser(username, password, publicKey string) error {
	// Try to get the user's secret from the local cache
	userSecret, found := k8s.GetSecretFromCache(username)
	if !found {
		return fmt.Errorf("user secret not found in cache")
	}
	secret := userSecret.(map[string]string)

	if secret["password"] != "" {
		if subtle.ConstantTimeCompare([]byte(password), []byte(secret["password"])) != 1 {
			return fmt.Errorf("password mismatch")
		}
	} else if secret["publicKey"] != "" {
		if err := authenticateWithPublicKey(publicKey, secret["publicKey"]); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("user secret must contain either a password or a publicKey")
	}

	return nil
}

func authenticateWithPublicKey(clientKey, storedPublicKey string) error {
	// Decode the stored public key
	decodedStoredPublicKey, err := base64.StdEncoding.DecodeString(storedPublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode stored public key: %v", err)
	}

	// Parse the public keys
	parsedClientKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(clientKey))
	if err != nil {
		return fmt.Errorf("failed to parse client public key: %v", err)
	}
	parsedStoredKey, _, _, _, err := ssh.ParseAuthorizedKey(decodedStoredPublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse stored public key: %v", err)
	}

	// Compare the public keys
	if subtle.ConstantTimeCompare(parsedClientKey.Marshal(), parsedStoredKey.Marshal()) != 1 {
		return fmt.Errorf("public key mismatch")
	}

	// Public key matches, proceed with authentication
	return nil
}
