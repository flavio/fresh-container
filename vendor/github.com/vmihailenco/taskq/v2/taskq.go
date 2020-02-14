package taskq

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/vmihailenco/taskq/v2/internal"
	"golang.org/x/time/rate"
)

func init() {
	SetLogger(log.New(os.Stderr, "taskq: ", log.LstdFlags|log.Lshortfile))
}

func SetLogger(logger *log.Logger) {
	internal.Logger = logger
}

// Factory is an interface that abstracts creation of new queues.
// It is implemented in subpackages memqueue, azsqs, and ironmq.
type Factory interface {
	RegisterQueue(*QueueOptions) Queue
	Range(func(Queue) bool)
	StartConsumers(context.Context) error
	StopConsumers() error
	Close() error
}

type Redis interface {
	Del(keys ...string) *redis.IntCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Pipelined(func(pipe redis.Pipeliner) error) ([]redis.Cmder, error)

	// Required by redislock
	Eval(script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd
	ScriptExists(scripts ...string) *redis.BoolSliceCmd
	ScriptLoad(script string) *redis.StringCmd
}

type Storage interface {
	Exists(key string) bool
}

type redisStorage struct {
	redis Redis
}

var _ Storage = (*redisStorage)(nil)

func newRedisStorage(redis Redis) redisStorage {
	return redisStorage{
		redis: redis,
	}
}

func (s redisStorage) Exists(key string) bool {
	if localCacheExists(key) {
		return true
	}

	val, err := s.redis.SetNX(key, "", 24*time.Hour).Result()
	if err != nil {
		return true
	}
	return !val
}

type RateLimiter interface {
	AllowRate(name string, limit rate.Limit) (delay time.Duration, allow bool)
}

//------------------------------------------------------------------------------

var (
	mu    sync.Mutex
	cache *simplelru.LRU
)

func localCacheExists(key string) bool {
	mu.Lock()
	defer mu.Unlock()

	if cache == nil {
		var err error
		cache, err = simplelru.NewLRU(128000, nil)
		if err != nil {
			panic(err)
		}
	}

	_, ok := cache.Get(key)
	if ok {
		return true
	}

	cache.Add(key, nil)
	return false
}
