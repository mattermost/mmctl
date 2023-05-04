package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// ===== Command output / return value types

// Nil represents the redis nil value
const Nil = redis.Nil

// Result generic interface
type Result interface {
	Int() int64
	String() string
	Bool() bool
	Duration() time.Duration
	Result() (int64, error)
	ResultString() (string, error)
	Multi() ([]string, error)
	MultiInterface() ([]interface{}, error)
	Err() error
	MapStringString() (map[string]string, error)
}

// ResultImpl generic interface
type ResultImpl struct {
	value           int64
	valueString     string
	valueBool       bool
	valueDuration   time.Duration
	err             error
	multi           []string
	multiInterface  []interface{}
	mapStringString map[string]string
}

// Int implementation
func (r *ResultImpl) Int() int64 {
	return r.value
}

// String implementation
func (r *ResultImpl) String() string {
	return r.valueString
}

// Bool implementation
func (r *ResultImpl) Bool() bool {
	return r.valueBool
}

// Duration implementation
func (r *ResultImpl) Duration() time.Duration {
	return r.valueDuration
}

// Err implementation
func (r *ResultImpl) Err() error {
	return r.err
}

// Result implementation
func (r *ResultImpl) Result() (int64, error) {
	return r.value, r.err
}

// ResultString implementation
func (r *ResultImpl) ResultString() (string, error) {
	return r.valueString, r.err
}

// Multi implementation
func (r *ResultImpl) Multi() ([]string, error) {
	return r.multi, r.err
}

// MultiInterface implementation
func (r *ResultImpl) MultiInterface() ([]interface{}, error) {
	return r.multiInterface, r.err
}

// MapStringString implementation
func (r *ResultImpl) MapStringString() (map[string]string, error) {
	return r.mapStringString, r.err
}

// Pipeline defines the interface of a redis pipeline
type Pipeline interface {
	LRange(key string, start, stop int64)
	LTrim(key string, start, stop int64)
	LLen(key string)
	HIncrBy(key string, field string, value int64)
	HLen(key string)
	Exec() ([]Result, error)
}

// PipelineImpl Wrapper
type PipelineImpl struct {
	wrapped redis.Pipeliner
}

// LRange schedules an lrange operation on this pipeline
func (p *PipelineImpl) LRange(key string, start, stop int64) {
	p.wrapped.LRange(context.TODO(), key, start, stop)
}

// LTrim schedules an ltrim operation on this pipeline
func (p *PipelineImpl) LTrim(key string, start, stop int64) {
	p.wrapped.LTrim(context.TODO(), key, start, stop)
}

// LLen schedules an llen operation on this pipeline
func (p *PipelineImpl) LLen(key string) {
	p.wrapped.LLen(context.TODO(), key)
}

// HIncrBy schedules an hincrby operation on this pipeline
func (p *PipelineImpl) HIncrBy(key string, field string, value int64) {
	p.wrapped.HIncrBy(context.TODO(), key, field, value)
}

// HLen schedules an HLen operation on this pipeline
func (p *PipelineImpl) HLen(key string) {
	p.wrapped.HLen(context.TODO(), key)
}

// Exec executes the pipeline
func (p *PipelineImpl) Exec() ([]Result, error) {
	res, err := p.wrapped.Exec(context.TODO())
	if err != nil {
		return nil, err
	}

	toRet := make([]Result, 0, len(res))
	for idx := range res {
		toRet = append(toRet, wrapResult(res[idx]))
	}

	return toRet, nil
}

// ====== Client

// Client interface which specifies the currently used subset of redis operations
type Client interface {
	ClusterMode() bool
	ClusterCountKeysInSlot(slot int) Result
	ClusterSlotForKey(key string) Result
	ClusterKeysInSlot(slot int, count int) Result
	Del(keys ...string) Result
	Exists(keys ...string) Result
	Get(key string) Result
	Set(key string, value interface{}, expiration time.Duration) Result
	Ping() Result
	Keys(pattern string) Result
	SMembers(key string) Result
	SIsMember(key string, member interface{}) Result
	SAdd(key string, members ...interface{}) Result
	SRem(key string, members ...interface{}) Result
	Incr(key string) Result
	Decr(key string) Result
	RPush(key string, values ...interface{}) Result
	LRange(key string, start, stop int64) Result
	LTrim(key string, start, stop int64) Result
	LLen(key string) Result
	Expire(key string, value time.Duration) Result
	TTL(key string) Result
	MGet(keys []string) Result
	SCard(key string) Result
	Eval(script string, keys []string, args ...interface{}) Result
	HIncrBy(key string, field string, value int64) Result
	HSet(key string, hashKey string, value interface{}) Result
	HGetAll(key string) Result
	Type(key string) Result
	Pipeline() Pipeline
}

// ClientImpl wrapps redis client
type ClientImpl struct {
	wrapped     redis.UniversalClient
	clusterMode bool
}

// ClusterMode returns true if the client is running in cluster mode
func (c *ClientImpl) ClusterMode() bool {
	return c.clusterMode
}

// ClusterSlotForKey returns the slot for the supplied key
func (c *ClientImpl) ClusterSlotForKey(key string) Result {
	res := c.wrapped.ClusterKeySlot(context.TODO(), key)
	return wrapResult(res)
}

// ClusterCountKeysInSlot returns the number of keys in slot
func (c *ClientImpl) ClusterCountKeysInSlot(slot int) Result {
	res := c.wrapped.ClusterCountKeysInSlot(context.TODO(), slot)
	return wrapResult(res)
}

// ClusterKeysInSlot returns all the keys in the supplied slot
func (c *ClientImpl) ClusterKeysInSlot(slot int, count int) Result {
	res := c.wrapped.ClusterGetKeysInSlot(context.TODO(), slot, count)
	return wrapResult(res)
}

// Del implements Del wrapper for redis
func (c *ClientImpl) Del(keys ...string) Result {
	res := c.wrapped.Del(context.TODO(), keys...)
	return wrapResult(res)
}

// Exists implements Exists wrapper for redis
func (c *ClientImpl) Exists(keys ...string) Result {
	res := c.wrapped.Exists(context.TODO(), keys...)
	return wrapResult(res)
}

// Get implements Get wrapper for redis
func (c *ClientImpl) Get(key string) Result {
	res := c.wrapped.Get(context.TODO(), key)
	return wrapResult(res)
}

// Set implements Set wrapper for redis
func (c *ClientImpl) Set(key string, value interface{}, expiration time.Duration) Result {
	res := c.wrapped.Set(context.TODO(), key, value, expiration)
	return wrapResult(res)
}

// Ping implements Ping wrapper for redis
func (c *ClientImpl) Ping() Result {
	res := c.wrapped.Ping(context.TODO())
	return wrapResult(res)
}

// Keys implements Keys wrapper for redis
func (c *ClientImpl) Keys(pattern string) Result {
	res := c.wrapped.Keys(context.TODO(), pattern)
	return wrapResult(res)
}

// SMembers implements SMembers wrapper for redis
func (c *ClientImpl) SMembers(key string) Result {
	res := c.wrapped.SMembers(context.TODO(), key)
	return wrapResult(res)
}

// SIsMember implements SIsMember wrapper for redis
func (c *ClientImpl) SIsMember(key string, member interface{}) Result {
	res := c.wrapped.SIsMember(context.TODO(), key, member)
	return wrapResult(res)
}

// SAdd implements SAdd wrapper for redis
func (c *ClientImpl) SAdd(key string, members ...interface{}) Result {
	res := c.wrapped.SAdd(context.TODO(), key, members...)
	return wrapResult(res)
}

// SRem implements SRem wrapper for redis
func (c *ClientImpl) SRem(key string, members ...interface{}) Result {
	res := c.wrapped.SRem(context.TODO(), key, members...)
	return wrapResult(res)
}

// Incr implements Incr wrapper for redis
func (c *ClientImpl) Incr(key string) Result {
	res := c.wrapped.Incr(context.TODO(), key)
	return wrapResult(res)
}

// Decr implements Decr wrapper for redis
func (c *ClientImpl) Decr(key string) Result {
	res := c.wrapped.Decr(context.TODO(), key)
	return wrapResult(res)
}

// RPush implements RPush wrapper for redis
func (c *ClientImpl) RPush(key string, values ...interface{}) Result {
	res := c.wrapped.RPush(context.TODO(), key, values...)
	return wrapResult(res)
}

// LRange implements LRange wrapper for redis
func (c *ClientImpl) LRange(key string, start, stop int64) Result {
	res := c.wrapped.LRange(context.TODO(), key, start, stop)
	return wrapResult(res)
}

// LTrim implements LTrim wrapper for redis
func (c *ClientImpl) LTrim(key string, start, stop int64) Result {
	res := c.wrapped.LTrim(context.TODO(), key, start, stop)
	return wrapResult(res)
}

// LLen implements LLen wrapper for redis
func (c *ClientImpl) LLen(key string) Result {
	res := c.wrapped.LLen(context.TODO(), key)
	return wrapResult(res)
}

// Expire implements Expire wrapper for redis
func (c *ClientImpl) Expire(key string, value time.Duration) Result {
	res := c.wrapped.Expire(context.TODO(), key, value)
	return wrapResult(res)
}

// TTL implements TTL wrapper for redis
func (c *ClientImpl) TTL(key string) Result {
	res := c.wrapped.TTL(context.TODO(), key)
	return wrapResult(res)
}

// MGet implements MGet wrapper for redis
func (c *ClientImpl) MGet(keys []string) Result {
	res := c.wrapped.MGet(context.TODO(), keys...)
	return wrapResult(res)
}

// SCard implements SCard wrapper for redis
func (c *ClientImpl) SCard(key string) Result {
	res := c.wrapped.SCard(context.TODO(), key)
	return wrapResult(res)
}

// Eval implements Eval wrapper for redis
func (c *ClientImpl) Eval(script string, keys []string, args ...interface{}) Result {
	res := c.wrapped.Eval(context.TODO(), script, keys, args...)
	return wrapResult(res)
}

// HIncrBy implements HIncrBy wrapper for redis
func (c *ClientImpl) HIncrBy(key string, field string, value int64) Result {
	res := c.wrapped.HIncrBy(context.TODO(), key, field, value)
	return wrapResult(res)
}

// HSet implements HSet wrapper for redis
func (c *ClientImpl) HSet(key string, hashKey string, value interface{}) Result {
	res := c.wrapped.HSet(context.TODO(), key, hashKey, value)
	return wrapResult(res)
}

// HGetAll implements HGetAll wrapper for redis
func (c *ClientImpl) HGetAll(key string) Result {
	res := c.wrapped.HGetAll(context.TODO(), key)
	return wrapResult(res)
}

// Type implements Type wrapper for redis
func (c *ClientImpl) Type(key string) Result {
	res := c.wrapped.Type(context.TODO(), key)
	return wrapResult(res)
}

// Pipeline implements Pipeline wrapper for redis
func (c *ClientImpl) Pipeline() Pipeline {
	res := c.wrapped.Pipeline()
	return &PipelineImpl{wrapped: res}
}

// NewClient returns new client implementation
func NewClient(options *UniversalOptions) (Client, error) {
	if options.ForceClusterMode {
		return &ClientImpl{wrapped: redis.NewClusterClient(options.toRedisClusterOpts()),
			clusterMode: true,
		}, nil
	}

	return &ClientImpl{wrapped: redis.NewUniversalClient(options.toRedisUniversalOpts()),
		clusterMode: len(options.Addrs) > 1 && options.MasterName == "",
	}, nil
}

func wrapResult(result interface{}) Result {
	if result == nil {
		return nil
	}
	switch v := result.(type) {
	case *redis.StatusCmd:
		return &ResultImpl{
			valueString: v.Val(),
			err:         v.Err(),
		}
	case *redis.IntCmd:
		return &ResultImpl{
			value: v.Val(),
			err:   v.Err(),
		}
	case *redis.StringCmd:
		return &ResultImpl{
			valueString: v.Val(),
			err:         v.Err(),
		}
	case *redis.StringSliceCmd:
		return &ResultImpl{
			err:   v.Err(),
			multi: v.Val(),
		}
	case *redis.BoolCmd:
		return &ResultImpl{
			valueBool: v.Val(),
			err:       v.Err(),
		}
	case *redis.DurationCmd:
		return &ResultImpl{
			valueDuration: v.Val(),
			err:           v.Err(),
		}
	case *redis.SliceCmd:
		return &ResultImpl{
			err:            v.Err(),
			multiInterface: v.Val(),
		}
	case *redis.Cmd:
		return &ResultImpl{
			err: v.Err(),
		}
	case *redis.StringStringMapCmd:
		return &ResultImpl{
			err:             v.Err(),
			mapStringString: v.Val(),
		}
	default:
		return nil
	}
}
