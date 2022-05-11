package redis

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	wredis "github.com/go-redis/redis/v8"
)

// UniversalOptions type used for redis package
// TODO(mredolatti): In order to avoid breaking the API, the struct now contains all original fields
// of go-redis' UniversalOptions struct, with our custom ones.
// Next time we bump toolkit's version, we should instead wrap or embed the original struct as well.
type UniversalOptions struct {

	// Original go-redis.UniversalOptions properties
	Addrs              []string
	DB                 int
	Dialer             func(ctx context.Context, network, addr string) (net.Conn, error)
	OnConnect          func(ctx context.Context, cn *wredis.Conn) error
	Username           string
	Password           string
	SentinelPassword   string
	MaxRetries         int
	MinRetryBackoff    time.Duration
	MaxRetryBackoff    time.Duration
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
	TLSConfig          *tls.Config
	MaxRedirects       int
	ReadOnly           bool
	RouteByLatency     bool
	RouteRandomly      bool
	MasterName         string

	// Custom properties
	ForceClusterMode bool
}

func (u *UniversalOptions) toRedisUniversalOpts() *wredis.UniversalOptions {
	return &wredis.UniversalOptions{
		Addrs:              u.Addrs,
		DB:                 u.DB,
		Dialer:             u.Dialer,
		OnConnect:          u.OnConnect,
		Username:           u.Username,
		Password:           u.Password,
		SentinelPassword:   u.SentinelPassword,
		MaxRetries:         u.MaxRetries,
		MinRetryBackoff:    u.MinRetryBackoff,
		MaxRetryBackoff:    u.MaxRetryBackoff,
		DialTimeout:        u.DialTimeout,
		ReadTimeout:        u.ReadTimeout,
		WriteTimeout:       u.WriteTimeout,
		PoolSize:           u.PoolSize,
		MinIdleConns:       u.MinIdleConns,
		MaxConnAge:         u.MaxConnAge,
		PoolTimeout:        u.PoolTimeout,
		IdleTimeout:        u.IdleTimeout,
		IdleCheckFrequency: u.IdleCheckFrequency,
		TLSConfig:          u.TLSConfig,
		MaxRedirects:       u.MaxRedirects,
		ReadOnly:           u.ReadOnly,
		RouteByLatency:     u.RouteByLatency,
		RouteRandomly:      u.RouteRandomly,
		MasterName:         u.MasterName,
	}
}

func (u *UniversalOptions) toRedisClusterOpts() *wredis.ClusterOptions {
	return &wredis.ClusterOptions{
		Addrs:              u.Addrs,
		MaxRedirects:       u.MaxRedirects,
		ReadOnly:           u.ReadOnly,
		RouteByLatency:     u.RouteByLatency,
		RouteRandomly:      u.RouteRandomly,
		Dialer:             u.Dialer,
		OnConnect:          u.OnConnect,
		Username:           u.Username,
		Password:           u.Password,
		MaxRetries:         u.MaxRetries,
		MinRetryBackoff:    u.MinRetryBackoff,
		MaxRetryBackoff:    u.MaxRetryBackoff,
		ReadTimeout:        u.ReadTimeout,
		DialTimeout:        u.DialTimeout,
		WriteTimeout:       u.WriteTimeout,
		PoolSize:           u.PoolSize,
		MinIdleConns:       u.MinIdleConns,
		MaxConnAge:         u.MaxConnAge,
		PoolTimeout:        u.PoolTimeout,
		IdleTimeout:        u.IdleTimeout,
		IdleCheckFrequency: u.IdleCheckFrequency,
		TLSConfig:          u.TLSConfig,
	}
}
