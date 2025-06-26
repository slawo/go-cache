package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type WriteLock struct {
	client    *redis.Client
	lockKey   string
	lockValue string
	unlocked  chan struct{}
}

func (l *WriteLock) Unlock() error {
	select {
	case <-l.unlocked:
		return nil // Already unlocked
	default:
		err := l.releaseLock(context.Background())
		if err != nil {
			return err
		}
		close(l.unlocked) // Close the channel to signal unlocked
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

func (l *WriteLock) releaseLock(ctx context.Context) error {
	// Use a Lua script to ensure atomicity of the unlock operation
	return l.client.Eval(ctx, unlockScript, []string{l.lockKey}, l.lockValue).Err()
}

const unlockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
`
