package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/slawo/go-cache/datastore"
)

func NewSynchroniser(ctx context.Context, opts ...SynchroniserOption) (*RedisSynchroniser, error) {
	o := SynchroniserOptions{}
	for _, opt := range opts {
		if err := opt.Apply(&o); err != nil {
			return nil, fmt.Errorf("synchroniser: failed to apply option: %w", err)
		}
	}

	if o.DSN == "" {
		return nil, fmt.Errorf("synchroniser: DSN cannot be empty")
	}
	if o.LockTimeoutSeconds == 0 {
		o.LockTimeoutSeconds = 6 // Default timeout of 6 seconds
	}

	client := redis.NewClient(&redis.Options{
		Addr:     o.DSN,
		Password: o.Password,
		DB:       o.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	mID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &RedisSynchroniser{
		client:             client,
		managerID:          mID.String(),
		lockTimeoutSeconds: o.LockTimeoutSeconds,
	}, nil
}

type RedisSynchroniser struct {
	client             *redis.Client
	managerID          string
	lockTimeoutSeconds int
}

func (r *RedisSynchroniser) GetWriteLock(ctx context.Context, lockID string) (datastore.DataWriteLock, error) {
	mID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(lockID) == "" || len(lockID) < 3 {
		return nil, datastore.ErrInvalidLockID
	}
	lockKey := "lock:" + lockID + ":write" // Ensure lockID is unique for write locks
	return NewWriteLock(ctx, r.client, lockKey, mID.String(), r.lockTimeoutSeconds)
}
