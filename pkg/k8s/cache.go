package k8s

import (
	"log"
	"time"

	"github.com/patrickmn/go-cache"
)

var localCache = cache.New(5*time.Minute, 10*time.Minute)

func GetSecretFromCache(username string) (interface{}, bool) {
	return localCache.Get(username)
}

func SetSecretInCache(username string, secretData map[string]string) {
	localCache.Set(username, secretData, cache.DefaultExpiration)
	log.Printf("Processed secret: %s, cache size: %d\n", username, localCache.ItemCount())
}

func DeleteSecretFromCache(username string) {
	localCache.Delete(username)
	log.Printf("Deleted secret: %s, cache size: %d\n", username, localCache.ItemCount())
}
