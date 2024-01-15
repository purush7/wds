package redis

import (
	"context"
	"time"
)

// acquireLock attempts to acquire a lock by setting a key with a specific value
func AcquireLock(lockKey, uniqueValue string, expiration time.Duration) (bool, error) {
	ctx := context.Background()

	// Use the SET command with NX and EX options to acquire the lock
	result, err := rdb.SetNX(ctx, lockKey, uniqueValue, expiration).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

// releaseLock releases the lock by deleting the key if its value matches the expected unique value
func ReleaseLock(lockKey, uniqueValue string) error {
	ctx := context.Background()

	// Use the Lua script to release the lock safely
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	_, err := rdb.Eval(ctx, script, []string{lockKey}, uniqueValue).Result()
	return err
}
