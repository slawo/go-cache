package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/slawo/go-cache/datastore"
)

func NewWriteLock(
	ctx context.Context,
	client *redis.Client,
	lockKey, lockValue string,
	timeoutSeconds int) (*WriteLock, error) {
	if client == nil {
		return nil, errors.New("redis client cannot be nil")
	}
	if lockKey == "" {
		return nil, errors.New("lock key cannot be empty")
	}
	if lockValue == "" {
		return nil, errors.New("lock value cannot be empty")
	}

	if len(lockKey) < 3 {
		return nil, errors.New("lock key must be at least 3 characters long")
	}

	l := &WriteLock{
		client:         client,
		lockKey:        lockKey,
		lockValue:      lockValue,
		stop:           make(chan struct{}),
		unlocked:       make(chan struct{}),
		timeoutSeconds: timeoutSeconds,
		tk:             time.NewTicker(time.Duration(timeoutSeconds) * time.Second / 2), // Default timeout of 10 seconds
	}

	if err := setLock(ctx, l); err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-l.unlocked:
				log.Default().Printf("lock %s has been unlocked", l.lockKey)
				return // Stop the goroutine when unlocked
			case <-l.stop:
				l.tk.Stop()
				err := releaseLock(ctx, l)
				if err != nil {
					log.Default().Printf("failed to release lock: %v", err)
				} // Close the channel to signal unlocked
				close(l.unlocked)
				return
			case <-l.tk.C:
				if err := setLock(ctx, l); err != nil {
					log.Default().Printf("failed to refresh lock: %v", err)
				} else {
					log.Default().Printf("lock refreshed: %s", l.lockKey)

				}
			}
		}
	}()

	return l, nil
}

type WriteLock struct {
	client         *redis.Client
	lockKey        string
	lockValue      string
	stop           chan struct{}
	unlocked       chan struct{}
	timeoutSeconds int
	tk             *time.Ticker
}

func (l *WriteLock) Unlock() (err error) {
	defer func() {
		<-l.unlocked
		if r := recover(); r != nil {
			err = errors.New("failed to unlock: " + r.(string))
		}
	}()
	select {
	case <-l.stop:
		return nil
	default:
		close(l.stop)
	}
	return nil
}

func (l *WriteLock) Unlocked() bool {
	select {
	case <-l.unlocked:
		return true // The lock has been released
	default:
	}
	return false // The lock is still held
}

func (l *WriteLock) WaitUnlocked() <-chan struct{} {
	return l.unlocked
}

func setLock(ctx context.Context, l *WriteLock) error {
	// Use a Lua script to ensure atomicity of the unlock operation
	fmt.Printf("Setting lock: %s with value: %s and timeout: %d seconds\n", l.lockKey, l.lockValue, l.timeoutSeconds)
	res := l.client.Eval(ctx, lockScript, []string{l.lockKey}, l.lockValue, l.timeoutSeconds)

	fmt.Printf("Setting lock: %s with value: %s result: %v, error: %v\n", l.lockKey, l.lockValue, res.Val(), res.Err())
	if res.Err() != nil {
		if errors.Is(res.Err(), redis.Nil) {
			fmt.Printf("Setting lock: %s with value: %s, FAILED (lock already held)\n", l.lockKey, l.lockValue)
			return fmt.Errorf("%v: %s", datastore.ErrLockAlreadyHeld, l.lockKey)
		}
		return res.Err()
	}
	if res.Val() != "OK" {
		fmt.Printf("Setting lock: %s with value: %s, FAILED\n", l.lockKey, l.lockValue)
		return fmt.Errorf("%v: %s", datastore.ErrLockAlreadyHeld, l.lockKey)
	}
	fmt.Printf("Setting lock: %s with value: %s, SUCCESS\n", l.lockKey, l.lockValue)
	return nil
}

func releaseLock(ctx context.Context, l *WriteLock) error {
	// Use a Lua script to ensure atomicity of the unlock operation
	fmt.Printf("Releasing lock: %s with value: %s\n", l.lockKey, l.lockValue)
	res := l.client.Eval(ctx, unlockScript, []string{l.lockKey}, l.lockValue)
	fmt.Printf("Releasing lock: %s, result: %v\n", l.lockKey, res.Val())
	if res.Err() != nil {
		return fmt.Errorf("failed to delease lock: %w", res.Err())
	}
	return nil
}

const lockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    return "OK"
else
    return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
end
`

const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`
