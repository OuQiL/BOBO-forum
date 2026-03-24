package svc

import (
	"testing"
	"time"

	"post/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisConnection(t *testing.T) {
	redisConf := redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
		Pass: "",
	}

	redisClient := redis.MustNewRedis(redisConf)
	assert.NotNil(t, redisClient)

	ok := redisClient.Ping()
	assert.True(t, ok, "Redis连接失败，请确保Redis服务已启动")
}

func TestRedisBasicOperations(t *testing.T) {
	redisConf := redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
		Pass: "",
	}

	redisClient := redis.MustNewRedis(redisConf)

	testKey := "test:redis:key"
	testValue := "test-value"

	err := redisClient.Set(testKey, testValue)
	assert.NoError(t, err)

	val, err := redisClient.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, testValue, val)

	_, err = redisClient.Del(testKey)
	assert.NoError(t, err)

	val, err = redisClient.Get(testKey)
	assert.NoError(t, err)
	assert.Empty(t, val)
}

func TestRedisSetWithExpiry(t *testing.T) {
	redisConf := redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
		Pass: "",
	}

	redisClient := redis.MustNewRedis(redisConf)

	testKey := "test:redis:expiry"
	testValue := "test-value-with-expiry"

	err := redisClient.Setex(testKey, testValue, 2)
	assert.NoError(t, err)

	val, err := redisClient.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, testValue, val)

	time.Sleep(3 * time.Second)

	val, err = redisClient.Get(testKey)
	assert.NoError(t, err)
	assert.Empty(t, val)
}

func TestRedisHSet(t *testing.T) {
	redisConf := redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
		Pass: "",
	}

	redisClient := redis.MustNewRedis(redisConf)

	testKey := "test:redis:hash"
	field1 := "field1"
	value1 := "value1"
	field2 := "field2"
	value2 := "value2"

	err := redisClient.Hset(testKey, field1, value1)
	assert.NoError(t, err)

	err = redisClient.Hset(testKey, field2, value2)
	assert.NoError(t, err)

	val1, err := redisClient.Hget(testKey, field1)
	assert.NoError(t, err)
	assert.Equal(t, value1, val1)

	val2, err := redisClient.Hget(testKey, field2)
	assert.NoError(t, err)
	assert.Equal(t, value2, val2)

	_, err = redisClient.Del(testKey)
	assert.NoError(t, err)
}

func TestRedisExpire(t *testing.T) {
	redisConf := redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
		Pass: "",
	}

	redisClient := redis.MustNewRedis(redisConf)

	testKey := "test:redis:expire"
	testValue := "test-value-expire"

	err := redisClient.Set(testKey, testValue)
	assert.NoError(t, err)

	err = redisClient.Expire(testKey, 2)
	assert.NoError(t, err)

	time.Sleep(3 * time.Second)

	val, err := redisClient.Get(testKey)
	assert.NoError(t, err)
	assert.Empty(t, val)
}

func TestRedisIncr(t *testing.T) {
	redisConf := redis.RedisConf{
		Host: "127.0.0.1:6379",
		Type: "node",
		Pass: "",
	}

	redisClient := redis.MustNewRedis(redisConf)

	testKey := "test:redis:incr"

	_, err := redisClient.Del(testKey)
	assert.NoError(t, err)

	val, err := redisClient.Incr(testKey)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = redisClient.Incr(testKey)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)

	val, err = redisClient.Incr(testKey)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), val)

	_, err = redisClient.Del(testKey)
	assert.NoError(t, err)
}

func TestServiceContextRedis(t *testing.T) {
	c := config.Config{
		Redis: redis.RedisConf{
			Host: "127.0.0.1:6379",
			Type: "node",
			Pass: "",
		},
	}

	svcCtx := NewServiceContext(c)
	assert.NotNil(t, svcCtx)
	assert.NotNil(t, svcCtx.Redis)

	testKey := "test:service:context:redis"
	testValue := "test-value"

	err := svcCtx.Redis.Set(testKey, testValue)
	assert.NoError(t, err)

	val, err := svcCtx.Redis.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, testValue, val)

	_, err = svcCtx.Redis.Del(testKey)
	assert.NoError(t, err)
}
