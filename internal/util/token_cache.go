package util

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/patrickmn/go-cache"
)

type TokenCache struct {
	cache *cache.Cache
}

func NewTokenCache(expiration time.Duration, maxItems int) *TokenCache {
	return &TokenCache{
		cache: cache.New(expiration, expiration*2),
	}
}

// GenerateToken generates a random token
func (tc *TokenCache) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Set stores a username with the given token
func (tc *TokenCache) Set(token, username string) {
	tc.cache.Set(token, username, cache.DefaultExpiration)
}

// Get retrieves the username associated with the token
func (tc *TokenCache) Get(token string) (string, bool) {
	username, found := tc.cache.Get(token)
	if !found {
		return "", false
	}
	return username.(string), true
}

// Delete removes a token from the cache
func (tc *TokenCache) Delete(token string) {
	tc.cache.Delete(token)
}
