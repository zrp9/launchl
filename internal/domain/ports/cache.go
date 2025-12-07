package ports

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zrp9/launchl/internal/domain"
)

type CacheOptions struct {
	TTL     time.Duration
	KeepTTL bool
	NX      bool
	XX      bool
}

type CacheOption func(*CacheOptions)

type ICacher interface {
	HealthCheck(ctx context.Context) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, val string, options ...CacheOption) error
	Delete(ctx context.Context, key string) error
	DeleteMulti(ctx context.Context, keys []string) error
}

type IStreameWriter interface {
	Write(ctx context.Context, fields domain.StreamEntry) (string, error)
	WriteEvent(ctx context.Context, msgType, target, src string, data json.RawMessage) (string, error)
	WriteJSON(ctx context.Context, fieldName string, payload []byte) (string, error)
}

type IStreamReader interface {
	ReadGroup(ctx context.Context, count int64) ([]domain.Message, error)
	Range(ctx context.Context, start, stop string) ([]domain.Message, error)
	RangeAll(ctx context.Context) ([]domain.Message, error)
	Ack(ctx context.Context, ids ...string) (int64, error)
	AckDel(ctx context.Context, ids ...string) (int64, error)
	ClaimIdle(ctx context.Context, minIdle time.Duration) (string, []domain.Message, error)
	Pending(ctx context.Context) ([]domain.Message, error)
}

type IStreamAdminer interface {
	CreateGroup(ctx context.Context) error
	TrimApprox(ctx context.Context) (int64, error)
	DeleteStream(ctx context.Context) error
	DeleteGroup(ctx context.Context) error
	DeleteConsumer(ctx context.Context) error
	StreamInfo(ctx context.Context) error
	GroupInfo(ctx context.Context) error
	ConsumerInfo(ctx context.Context) error
}
