package redis

import (
	"fmt"
	"strings"
	"time"
)

// PrefixedRedisClient struct
type PrefixedRedisClient struct {
	prefix string
	client Client
}

// Prefix returns the prefix attached to the client
func (p *PrefixedRedisClient) Prefix() string {
	return p.prefix
}

// ClusterMode returns true if the client is working in cluster mode
func (p *PrefixedRedisClient) ClusterMode() bool {
	return p.client.ClusterMode()
}

// ClusterSlotForKey returns the slot for the supplied key
func (p *PrefixedRedisClient) ClusterSlotForKey(key string) (int64, error) {
	return p.client.ClusterSlotForKey(withPrefix(p.prefix, key)).Result()
}

// ClusterCountKeysInSlot returns the number of keys in slot
func (p *PrefixedRedisClient) ClusterCountKeysInSlot(slot int) (int64, error) {
	return p.client.ClusterCountKeysInSlot(slot).Result()
}

// ClusterKeysInSlot returns all the keys in the supplied slot
func (p *PrefixedRedisClient) ClusterKeysInSlot(slot int, count int) ([]string, error) {
	keys, err := p.client.ClusterKeysInSlot(slot, count).Multi()
	if err != nil {
		return nil, err
	}

	woPrefix := make([]string, len(keys))
	for index, key := range keys {
		woPrefix[index] = withoutPrefix(p.prefix, key)
	}
	return woPrefix, nil
}

// Get wraps around redis get method by adding prefix and returning string and error directly
func (p *PrefixedRedisClient) Get(key string) (string, error) {
	return p.client.Get(withPrefix(p.prefix, key)).ResultString()
}

// Set wraps around redis get method by adding prefix and returning error directly
func (p *PrefixedRedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return p.client.Set(withPrefix(p.prefix, key), value, expiration).Err()
}

// Keys wraps around redis keys method by adding prefix and returning []string and error directly
func (p *PrefixedRedisClient) Keys(pattern string) ([]string, error) {
	keys, err := p.client.Keys(withPrefix(p.prefix, pattern)).Multi()
	if err != nil {
		return nil, err
	}

	woPrefix := make([]string, len(keys))
	for index, key := range keys {
		woPrefix[index] = withoutPrefix(p.prefix, key)
	}
	return woPrefix, nil

}

// Del wraps around redis del method by adding prefix and returning int64 and error directly
func (p *PrefixedRedisClient) Del(keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = withPrefix(p.prefix, k)
	}
	return p.client.Del(prefixedKeys...).Result()
}

// SMembers returns a slice with all the members of a set
func (p *PrefixedRedisClient) SMembers(key string) ([]string, error) {
	return p.client.SMembers(withPrefix(p.prefix, key)).Multi()
}

// SIsMember returns true if members is in the set
func (p *PrefixedRedisClient) SIsMember(key string, member interface{}) bool {
	return p.client.SIsMember(withPrefix(p.prefix, key), member).Bool()
}

// SAdd adds new members to a set
func (p *PrefixedRedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	return p.client.SAdd(withPrefix(p.prefix, key), members...).Result()
}

// SRem removes members from a set
func (p *PrefixedRedisClient) SRem(key string, members ...interface{}) (int64, error) {
	return p.client.SRem(withPrefix(p.prefix, key), members...).Result()
}

// Exists returns true if a key exists in redis
func (p *PrefixedRedisClient) Exists(keys ...string) (int64, error) {
	prefixedKeys := make([]string, len(keys))
	for i, k := range keys {
		prefixedKeys[i] = withPrefix(p.prefix, k)
	}
	val, err := p.client.Exists(prefixedKeys...).Result()
	return val, err
}

// Incr increments a key. Sets it in one if it doesn't exist
func (p *PrefixedRedisClient) Incr(key string) (int64, error) {
	return p.client.Incr(withPrefix(p.prefix, key)).Result()
}

// Decr increments a key. Sets it in one if it doesn't exist
func (p *PrefixedRedisClient) Decr(key string) (int64, error) {
	return p.client.Decr(withPrefix(p.prefix, key)).Result()
}

// RPush insert all the specified values at the tail of the list stored at key
func (p *PrefixedRedisClient) RPush(key string, values ...interface{}) (int64, error) {
	return p.client.RPush(withPrefix(p.prefix, key), values...).Result()
}

// LRange Returns the specified elements of the list stored at key
func (p *PrefixedRedisClient) LRange(key string, start, stop int64) ([]string, error) {
	return p.client.LRange(withPrefix(p.prefix, key), start, stop).Multi()
}

// LTrim Trim an existing list so that it will contain only the specified range of elements specified
func (p *PrefixedRedisClient) LTrim(key string, start, stop int64) error {
	return p.client.LTrim(withPrefix(p.prefix, key), start, stop).Err()
}

// LLen Returns the length of the list stored at key
func (p *PrefixedRedisClient) LLen(key string) (int64, error) {
	return p.client.LLen(withPrefix(p.prefix, key)).Result()
}

// Expire set expiration time for particular key
func (p *PrefixedRedisClient) Expire(key string, value time.Duration) bool {
	return p.client.Expire(withPrefix(p.prefix, key), value).Bool()
}

// TTL for particular key
func (p *PrefixedRedisClient) TTL(key string) time.Duration {
	return p.client.TTL(withPrefix(p.prefix, key)).Duration()
}

// MGet fetchs multiple results
func (p *PrefixedRedisClient) MGet(keys []string) ([]interface{}, error) {
	keysWithPrefix := make([]string, 0)
	for _, key := range keys {
		keysWithPrefix = append(keysWithPrefix, withPrefix(p.prefix, key))
	}
	return p.client.MGet(keysWithPrefix).MultiInterface()
}

// SCard implements SCard wrapper for redis
func (p *PrefixedRedisClient) SCard(key string) (int64, error) {
	return p.client.SCard(withPrefix(p.prefix, key)).Result()
}

// Eval implements Eval wrapper for redis
func (p *PrefixedRedisClient) Eval(script string, keys []string, args ...interface{}) error {
	return p.client.Eval(script, keys, args...).Err()
}

// HIncrBy implements HIncrBy wrapper for redis
func (p *PrefixedRedisClient) HIncrBy(key string, field string, value int64) (int64, error) {
	return p.client.HIncrBy(withPrefix(p.prefix, key), field, value).Result()
}

// HSet implements HGetAll wrapper for redis
func (p *PrefixedRedisClient) HSet(key string, hashKey string, value interface{}) error {
	return p.client.HSet(withPrefix(p.prefix, key), hashKey, value).Err()
}

// HGetAll implements HGetAll wrapper for redis
func (p *PrefixedRedisClient) HGetAll(key string) (map[string]string, error) {
	return p.client.HGetAll(withPrefix(p.prefix, key)).MapStringString()
}

// Type implements Type wrapper for redis with prefix
func (p *PrefixedRedisClient) Type(key string) (string, error) {
	return p.client.Type(withPrefix(p.prefix, key)).ResultString()
}

// Pipeline wrapper
func (p *PrefixedRedisClient) Pipeline() Pipeline {
	return &PrefixedPipeline{wrapped: p.client.Pipeline(), prefix: p.prefix}
}

// NewPrefixedRedisClient returns a new Prefixed Redis Client
func NewPrefixedRedisClient(redisClient Client, prefix string) (*PrefixedRedisClient, error) {
	return &PrefixedRedisClient{
		client: redisClient,
		prefix: prefix,
	}, nil
}

// PrefixedPipeline adds a prefix to all pipelined operations involving keys
type PrefixedPipeline struct {
	prefix  string
	wrapped Pipeline
}

// LRange schedules an lrange operation on this pipeline
func (p *PrefixedPipeline) LRange(key string, start, stop int64) {
	p.wrapped.LRange(withPrefix(p.prefix, key), start, stop)
}

// LTrim schedules an ltrim operation on this pipeline
func (p *PrefixedPipeline) LTrim(key string, start, stop int64) {
	p.wrapped.LTrim(withPrefix(p.prefix, key), start, stop)
}

// LLen schedules an llen operation on this pipeline
func (p *PrefixedPipeline) LLen(key string) {
	p.wrapped.LLen(withPrefix(p.prefix, key))
}

// HIncrBy schedules an hincrby operation on this pipeline
func (p *PrefixedPipeline) HIncrBy(key string, field string, value int64) {
	p.wrapped.HIncrBy(withPrefix(p.prefix, key), field, value)
}

// HLen schedules an HLen operation on this pipeline
func (p *PrefixedPipeline) HLen(key string) {
	p.wrapped.HLen(withPrefix(p.prefix, key))
}

// Exec executes the pipeline
func (p *PrefixedPipeline) Exec() ([]Result, error) {
	return p.wrapped.Exec()
}

// withPrefix adds a prefix to the key if the prefix supplied has a length greater than 0
func withPrefix(prefix string, key string) string {
	if len(prefix) > 0 {
		return fmt.Sprintf("%s.%s", prefix, key)
	}
	return key
}

// withoutPrefix removes the prefix from a key if the prefix has a length greater than 0
func withoutPrefix(prefix string, key string) string {
	if len(prefix) > 0 {
		return strings.Replace(key, fmt.Sprintf("%s.", prefix), "", 1)
	}
	return key
}
