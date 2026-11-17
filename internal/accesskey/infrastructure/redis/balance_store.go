package redis

import (
	"context"
	"fmt"
	"strconv"

	redisclient "github.com/redis/go-redis/v9"
)

var initializeBalanceScript = redisclient.NewScript(`
local current = redis.call('GET', KEYS[1])
if current then
    return current
end
redis.call('SET', KEYS[1], ARGV[1])
return ARGV[1]
`)

type BalanceStore struct {
	client *redisclient.Client
}

func NewBalanceStore(client *redisclient.Client) *BalanceStore {
	return &BalanceStore{client: client}
}

func (s *BalanceStore) GetOrInitialize(ctx context.Context, userID, initialBalance int64) (int64, error) {
	value, err := initializeBalanceScript.Run(
		ctx,
		s.client,
		[]string{balanceKey(userID)},
		strconv.FormatInt(initialBalance, 10),
	).Int64()
	if err != nil {
		return 0, fmt.Errorf("get or initialize user balance: %w", err)
	}
	return value, nil
}

func balanceKey(userID int64) string {
	return "balance:" + strconv.FormatInt(userID, 10)
}
