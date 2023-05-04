package redis

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/splitio/go-split-commons/v4/conf"
	"github.com/splitio/go-toolkit/v5/logging"
	"github.com/splitio/go-toolkit/v5/redis"
)

// Redis initialization erorrs
var (
	ErrInvalidConf           = errors.New("incompatible configuration of redis, Sentinel and Cluster cannot be enabled at the same time")
	ErrSentinelNoMaster      = errors.New("Missing redis sentinel master name")
	ErrClusterInvalidHashtag = errors.New("hashtag must be wrapped in '{', '}', and be at least 3 characters long")
)

// NewRedisClient returns a new Prefixed Redis Client
func NewRedisClient(config *conf.RedisConfig, logger logging.LoggerInterface) (*redis.PrefixedRedisClient, error) {
	prefix := config.Prefix

	if len(config.SentinelAddresses) > 0 && len(config.ClusterNodes) > 0 {
		return nil, ErrInvalidConf
	}

	universalOptions := &redis.UniversalOptions{
		Password:     config.Password,
		Username:     config.Username,
		DB:           config.Database,
		TLSConfig:    config.TLSConfig,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
	}

	if len(config.SentinelAddresses) > 0 {
		if config.SentinelMaster == "" {
			return nil, ErrSentinelNoMaster
		}

		universalOptions.MasterName = config.SentinelMaster
		universalOptions.Addrs = config.SentinelAddresses
	} else {
		if len(config.ClusterNodes) > 0 {
			var keyHashTag = "{SPLITIO}"

			if config.ClusterKeyHashTag != "" {
				keyHashTag = config.ClusterKeyHashTag
				if len(keyHashTag) < 3 ||
					string(keyHashTag[0]) != "{" ||
					string(keyHashTag[len(keyHashTag)-1]) != "}" ||
					strings.Count(keyHashTag, "{") != 1 ||
					strings.Count(keyHashTag, "}") != 1 {
					return nil, ErrClusterInvalidHashtag
				}
			}

			prefix = keyHashTag + prefix
			universalOptions.Addrs = config.ClusterNodes
			universalOptions.ForceClusterMode = true // to enable auto-discovery of nodes when providing only one
		} else {
			universalOptions.Addrs = []string{fmt.Sprintf("%s:%d", config.Host, config.Port)}
		}
	}

	rClient, err := redis.NewClient(universalOptions)

	if err != nil {
		return nil, fmt.Errorf("error constructing wrapped redis client: %w", err)
	}

	if res := rClient.Ping(); res.Err() != nil {
		if asOpErr, ok := res.Err().(*net.OpError); ok {
			return nil, asOpErr // don't wrap it to conserve type
		}
		return nil, fmt.Errorf("couldn't connect to redis: %w", err)
	}

	return redis.NewPrefixedRedisClient(rClient, prefix)
}
