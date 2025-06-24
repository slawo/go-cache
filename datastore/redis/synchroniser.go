package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/slawo/go-cache/datastore"
)

func NewSynchroniser(ctx context.Context, dsn string, password string, db int) (*RedisSynchroniser, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     dsn,
		Password: password, // no password set
		DB:       db,       // use default DB
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
		defaultLockTimeout: 10 * time.Second, // Default lock timeout
	}, nil
}

type RedisSynchroniser struct {
	client             *redis.Client
	managerID          string
	defaultLockTimeout time.Duration
}

func (r *RedisSynchroniser) GetWriteLock(ctx context.Context, lockID string) (datastore.DataWriteLock, error) {
	if strings.TrimSpace(lockID) == "" || len(lockID) < 3 {
		return nil, datastore.ErrInvalidLockID
	}
	lockKey := "lock:" + lockID + ":write" // Ensure lockID is unique for write locks
	res := r.client.SetArgs(ctx, lockKey, r.managerID, redis.SetArgs{
		TTL:  r.defaultLockTimeout,
		Mode: "NX",
		Get:  true,
	})

	if res.Err() != nil && res.Err() != redis.Nil {
		return nil, res.Err()
	} else if res.Val() != "" {

		return nil, fmt.Errorf("%w: %s", datastore.ErrLockAlreadyHeld, lockID)
	}

	return &WriteLock{
		client:    r.client,
		lockKey:   lockKey,
		lockValue: r.managerID,
		unlocked:  make(chan struct{}),
	}, nil
}

// func (r *RedisSynchroniser) aquireLock(ctx context.Context, lockKey string) (*WriteLock, error) {

// 	// err := r.client.Set(ctx, "key", "value", 0).Err()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// val, err := r.client.Get(ctx, "key").Result()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// fmt.Println("key", val)

// 	// val2, err := r.client.Get(ctx, "key2").Result()
// 	// if err == redis.Nil {
// 	// 	fmt.Println("key2 does not exist")
// 	// } else if err != nil {
// 	// 	panic(err)
// 	// } else {
// 	// 	fmt.Println("key2", val2)
// 	// }
// 	// return nil, nil // Placeholder for actual lock acquisition logic
// }
