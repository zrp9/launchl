// Package valkaree exposes a inteface to valkey for caching and streaming
package valkaree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/domain/ports"
)

var ErrEmptyCache = errors.New("no results found for given key")

func NewClient(ctx context.Context, cfg config.ValkeyCfg) (valkey.Client, error) {
	log.Printf("valkey host %v, valkey port %v\n", cfg.Host, cfg.Port)
	if cfg.Host == "" || cfg.Port == "" {
		return nil, errors.New("valkey host and port must be defined")
	}

	log.Printf("valkey using %s:%s", cfg.Host, cfg.Port)
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
	})

	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pong, err := client.Do(pingCtx, client.B().Ping().Build()).ToString()
	if err != nil {
		client.Close()
		return nil, err
	}

	fmt.Println("valkey connected: ", pong)

	return client, nil
}

type Cache struct {
	client valkey.Client
}

func NewCache(c valkey.Client) Cache {
	return Cache{
		client: c,
	}
}

func (v Cache) HealthCheck(ctx context.Context) error {
	status, err := v.client.Do(ctx, v.client.B().Ping().Build()).ToString()
	if err != nil {
		return fmt.Errorf("health check failed cache down, %w", err)
	}

	if status != "PONG" {
		return fmt.Errorf("health check failed cache down")
	}

	return nil
}

func (v Cache) Set(ctx context.Context, key, value string, opts ...ports.CacheOption) error {
	cfg := &ports.CacheOptions{}

	for _, opt := range opts {
		opt(cfg)
	}

	b := v.client.B().
		Set().
		Key(key).
		Value(value)

	if cfg.TTL > 0 {
		b.Ex(cfg.TTL)
	}
	if cfg.KeepTTL {
		b.Keepttl()
	}
	if cfg.NX {
		b.Nx()
	}
	if cfg.XX {
		b.Xx()
	}

	return v.client.Do(ctx, b.Build()).Error()
	//return v.client.Do(ctx, v.client.B().Set().Key(key).Value(value).Build()).Error()
}

func (v Cache) Get(ctx context.Context, key string) (string, error) {
	val, err := v.client.Do(ctx, v.client.B().Get().Key(key).Build()).ToString()
	if err != nil {
		if err == valkey.Nil {
			return "", ErrEmptyCache
		}
		return "", err
	}

	return val, nil
}

func (v Cache) Delete(ctx context.Context, key string) error {
	return v.client.Do(ctx, v.client.B().Del().Key(key).Build()).Error()
}

func (v Cache) DeleteMulti(ctx context.Context, keys []string) error {
	return v.client.Do(ctx, v.client.B().Del().Key(keys...).Build()).Error()
}

type Stream struct {
	client       valkey.Client
	Key          string
	threshold    int64
	writeRetries int64
	log          crane.Zlogrus
}

type writer struct {
	s            *Stream
	MaxLenApprox int64
	AckedOnly    bool
}

type reader struct {
	s        *Stream
	Group    string
	Consumer string
	Block    time.Duration
	count    int64
}

type admin struct {
	s *Stream
	// Group    string
	// Consumer string
}

type master struct {
	writer
	reader
	admin
}

type Event struct {
	DataType  string
	EventType string
	Target    string
	Src       string
}

type StreamWriter interface {
	Write(ctx context.Context, fields domain.StreamEntry) (string, error)
	WriteEvent(ctx context.Context, event Event, payload json.RawMessage) (string, error)
	WriteJSON(ctx context.Context, fieldName string, payload []byte) (string, error)
}

type StreamReader interface {
	ReadGroup(ctx context.Context, count int64) ([]domain.Message, error)
	Range(ctx context.Context, start, stop string) ([]domain.Message, error)
	RangeAll(ctx context.Context) ([]domain.Message, error)
	Ack(ctx context.Context, ids ...string) (int64, error)
	AckDel(ctx context.Context, ids ...string) (int64, error)
	ClaimIdle(ctx context.Context, minIdle time.Duration) (string, []domain.Message, error)
	Pending(ctx context.Context) ([]domain.Message, error)
	Count() int64
}

// StreamAdmin update this so it can create a stream then enable delete stream
type StreamAdmin interface {
	CreateGroup(ctx context.Context, group string) error
	TrimApprox(ctx context.Context) (int64, error)
	//DeleteStream(ctx context.Context) error
	DeleteGroup(ctx context.Context, group string) error
	DeleteConsumer(ctx context.Context, group, consumer string) error
	StreamInfo(ctx context.Context) error
	GroupInfo(ctx context.Context) error
	ConsumerInfo(ctx context.Context, group, consumer string) error
}

type StreamMaster interface {
	StreamWriter
	StreamReader
	StreamAdmin
}

func NewStream(client valkey.Client, key string, writeRetries, threshold int64, log crane.Zlogrus) *Stream {
	return &Stream{
		client:       client,
		Key:          key,
		writeRetries: writeRetries,
		threshold:    threshold,
		log:          log,
	}
}

func (s *Stream) Writer() StreamWriter { return writer{s: s} }

func (s *Stream) Reader(group, consumer string, block time.Duration, count int64) StreamReader {
	return reader{
		s:        s,
		Group:    group,
		Consumer: consumer,
		Block:    block,
		count:    count,
	}
}

func (s *Stream) Admin() StreamAdmin {
	return admin{
		s: s,
		// Group:    group,
		// Consumer: consumer,
	}
}

func (s *Stream) Master(group, consumer string, block time.Duration, count int64) StreamMaster {
	return &master{
		writer: writer{s: s},
		reader: reader{
			s:        s,
			Group:    group,
			Consumer: consumer,
			Block:    block,
			count:    count,
		},
		admin: admin{
			s: s,
			// Group:    group,
			// Consumer: consumer,
		},
	}
}

func (r admin) CreateGroup(ctx context.Context, group string) error {
	res := r.s.client.Do(ctx, r.s.client.B().XgroupCreate().Key(r.s.Key).Group(group).Id("$").Mkstream().Build())

	if err := res.Error(); err != nil {
		var vErr *valkey.ValkeyError
		if errors.As(err, &vErr) && vErr.IsBusyGroup() {
			// group already exists
			return nil
		}
		log.Printf("error creating group %v", err)
		return err
	}

	return nil
}

func (w writer) Write(ctx context.Context, fields domain.StreamEntry) (string, error) {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cmd := w.s.client.B().Xadd().Key(w.s.Key).Maxlen().Almost().Threshold(w.s.Threshold()).Id("*").FieldValue().FieldValueIter(func(yield func(string, string) bool) {
		for _, k := range keys {
			if !yield(k, fields[k]) {
				return
			}
		}
	})

	return w.s.client.Do(ctx, cmd.Build()).ToString()
}

func (w writer) WriteEvent(ctx context.Context, event Event, payload json.RawMessage) (string, error) {
	log.Println("")
	log.Printf("stream %v write event: target: %v,kind: %v, src: %v payload: %v", w.s.Key, event.Target, event.EventType, event.Src, payload)
	log.Println("")
	cmd := w.s.client.B().Xadd().Key(w.s.Key).Maxlen().Almost().Threshold(w.s.Threshold()).Id("*").FieldValue().FieldValueIter(func(yield func(string, string) bool) {
		if !yield("jid", uuid.NewString()) {
			return
		}

		if !yield("eventType", event.EventType) {
			return
		}

		if !yield("target", event.Target) {
			return
		}

		if !yield("source", event.Src) {
			return
		}

		if !yield("retryLimit", strconv.FormatInt(w.s.writeRetries, 10)) {
			return
		}

		if !yield("payload", valkey.BinaryString(payload)) {
			return
		}
	})

	return w.s.client.Do(ctx, cmd.Build()).ToString()
}

func (w writer) WriteJSON(ctx context.Context, jsonField string, payload []byte) (string, error) {
	cmd := w.s.client.B().Xadd().Key(w.s.Key).Maxlen().Almost().Threshold(w.s.Threshold()).Id("*").FieldValue().FieldValueIter(func(yield func(string, string) bool) {
		_ = yield(jsonField, valkey.BinaryString(payload))
	})
	return w.s.client.Do(ctx, cmd.Build()).ToString()
}

func (r reader) ReadGroup(ctx context.Context, count int64) ([]domain.Message, error) {
	log.Printf("reading from stream %v", r.s.Key)
	log.Printf("reading to group %v", r.Group)
	var c int64
	if count == 0 {
		c = r.count
	}

	cmd := r.s.client.B().Xreadgroup().Group(r.Group, r.Consumer).Count(c).Block(r.Block.Milliseconds()).Streams().Key(r.s.Key).Id(">").Build()
	res := r.s.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		log.Printf("error reading group %v", err)
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}

	msgs, err := r.s.parseXRead(res)
	if err != nil {
		return nil, err
	}

	log.Printf("read group success")
	return msgs, nil
}

func (r reader) Range(ctx context.Context, start, stop string) ([]domain.Message, error) {
	res := r.s.client.Do(ctx, r.s.client.B().Xrange().Key(r.s.Key).Start(start).End(stop).Build())

	if err := res.Error(); err != nil {
		return nil, err
	}

	msgs, err := r.s.parseXRange(res)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (r reader) RangeAll(ctx context.Context) ([]domain.Message, error) {
	res := r.s.client.Do(ctx, r.s.client.B().Xrange().Key(r.s.Key).Start("-").End("+").Build())

	if err := res.Error(); err != nil {
		return nil, err
	}

	msgs, err := r.s.parseXRange(res)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (r reader) Ack(ctx context.Context, ids ...string) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("atleast one id is required to acknowledge")
	}
	cmd := r.s.client.B().Xack().Key(r.s.Key).Group(r.Group).Id(ids...).Build()
	return r.s.client.Do(ctx, cmd).ToInt64()
}

func (r reader) AckDel(ctx context.Context, ids ...string) (int64, error) {
	if len(ids) == 0 {
		return 0, errors.New("atleast one id is required to acknowledge")
	}
	cmd := r.s.client.B().Xackdel().Key(r.s.Key).Group(r.Group).Ids().Numids(int64(len(ids))).Id(ids...).Build()

	return r.s.client.Do(ctx, cmd).ToInt64()
}

func (r reader) ClaimIdle(ctx context.Context, minIdle time.Duration) (next string, msgs []domain.Message, err error) {
	cmd := r.s.client.B().Xautoclaim().Key(r.s.Key).Group(r.Group).Consumer(r.Consumer).MinIdleTime(r.s.toMs(minIdle)).Start("0-0").Count(r.count).Build()

	res := r.s.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		return "", nil, err
	}

	return r.s.parseXAutoClaim(res)
}

func (r reader) Pending(ctx context.Context) ([]domain.Message, error) {
	cmd := r.s.client.B().Xreadgroup().Group(r.Group, r.Consumer).Streams().Key(r.s.Key).Id("0").Build()
	res := r.s.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		return nil, err
	}
	return r.s.parseXRead(res)
}

func (r reader) Count() int64 {
	return r.count
}

func (r admin) TrimApprox(ctx context.Context) (int64, error) {
	cmd := r.s.client.B().Xtrim().Key(r.s.Key).Maxlen().Almost().Threshold(r.s.Threshold()).Acked().Build()
	return r.s.client.Do(ctx, cmd).ToInt64()
}

// func (r admin) DeleteStream(ctx context.Context) error {
// 	cmd := r.s.client.B().Del().Key(r.s.Key).Build()
// 	res := r.s.client.Do(ctx, cmd)
// 	return res.Error()
// }

func (r admin) DeleteGroup(ctx context.Context, group string) error {
	cmd := r.s.client.B().XgroupDestroy().Key(r.s.Key).Group(group).Build()
	if err := r.s.client.Do(ctx, cmd).Error(); err != nil {
		r.s.log.MustDebug(fmt.Sprintf("error occurred while deleting stream group %v", err))
		return err
	}
	return nil
}

func (r admin) DeleteConsumer(ctx context.Context, group, consumer string) error {
	cmd := r.s.client.B().XgroupDelconsumer().Key(r.s.Key).Group(group).Consumername(consumer).Build()
	if err := r.s.client.Do(ctx, cmd).Error(); err != nil {
		r.s.log.MustDebug(fmt.Sprintf("error deleting stream consumer %v", err))
		return err
	}

	return nil
}

func (r admin) StreamInfo(ctx context.Context) error {
	cmd := r.s.client.B().XinfoStream().Key(r.s.Key).Build()
	res := r.s.client.Do(ctx, cmd)
	log.Printf("x info stream result %#v", res)
	return nil
}

func (r admin) GroupInfo(ctx context.Context) error {
	cmd := r.s.client.B().XinfoGroups().Key(r.s.Key).Build()
	res := r.s.client.Do(ctx, cmd)
	log.Printf("x info groups result %#v", res)
	return nil
}

func (r admin) ConsumerInfo(ctx context.Context, group, consumer string) error {
	cmd := r.s.client.B().XinfoConsumers().Key(r.s.Key).Group(group).Build()
	res := r.s.client.Do(ctx, cmd)
	log.Printf("x info consumer result %#v", res)
	return nil
}

// Parses Xread / XreadGroup -> map[string][]XRangeSlices where keys are stream names
func (s *Stream) parseXRead(res valkey.ValkeyResult) ([]domain.Message, error) {
	msgs := make([]domain.Message, 0)
	streams, err := res.AsXReadSlices()
	if err != nil {
		return nil, err
	}

	// stream, entries can loop over multiple streams
	for _, entries := range streams {
		for _, e := range entries {
			msgs = append(msgs, domain.Message{
				ID:     e.ID,
				Values: s.parseFields(e.FieldValues),
			})
		}
	}

	return msgs, nil
}

// XRANGE / XREVRANGE → []XRangeSlice
func (s *Stream) parseXRange(res valkey.ValkeyResult) ([]domain.Message, error) {
	slices, err := res.AsXRangeSlices()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]domain.Message, 0, len(slices))
	for _, e := range slices {
		out = append(out, domain.Message{ID: e.ID, Values: s.parseFields(e.FieldValues)})
	}
	return out, nil
}

// XAUTOCLAIM → [ nextStartId, [ entries... ] , (opt: deleted count) ]
func (s *Stream) parseXAutoClaim(res valkey.ValkeyResult) (next string, msgs []domain.Message, err error) {
	arr, err := res.ToArray()
	if err != nil {
		return "", nil, err
	}
	if len(arr) < 2 {
		return "", nil, fmt.Errorf("xautoclaim: unexpected reply len %d", len(arr))
	}

	next, _ = arr[0].ToString()

	entrySlices, err := arr[1].AsXRangeSlices()
	if err != nil {
		// some builds need manual parse; fallback:
		entriesArr, err2 := arr[1].ToArray()
		if err2 != nil {
			return "", nil, err
		}
		entrySlices = make([]valkey.XRangeSlice, 0, len(entriesArr))
		for _, it := range entriesArr {
			idVals, _ := it.ToArray() // [id, [k,v,...]]
			id, _ := idVals[0].ToString()
			kv, _ := idVals[1].ToArray()
			pairs := make([]valkey.XRangeFieldValue, 0, len(kv)/2)
			for i := 0; i+1 < len(kv); i += 2 {
				k, _ := kv[i].ToString()
				v, _ := kv[i+1].ToString()
				pairs = append(pairs, valkey.XRangeFieldValue{Field: k, Value: v})
			}
			entrySlices = append(entrySlices, valkey.XRangeSlice{ID: id, FieldValues: pairs})
		}
	}
	msgs = make([]domain.Message, 0, len(entrySlices))
	for _, e := range entrySlices {
		msgs = append(msgs, domain.Message{ID: e.ID, Values: s.parseFields(e.FieldValues)})
	}
	return next, msgs, nil
}

func (s *Stream) parseFields(fieldValues []valkey.XRangeFieldValue) []domain.Field {
	fields := make([]domain.Field, 0, len(fieldValues))
	for _, fv := range fieldValues {
		fields = append(fields, domain.Field{
			Name:  fv.Field,
			Value: fv.Value,
		})
	}

	return fields
}

func (s *Stream) toMs(d time.Duration) string {
	return strconv.FormatInt(d.Milliseconds(), 10)
}

func (s *Stream) Threshold() string {
	return strconv.FormatInt(s.threshold, 10)
}

// compile checks force compile err if type doesnt implement interface
var _ StreamWriter = (*writer)(nil)
var _ StreamReader = (*reader)(nil)
var _ StreamAdmin = (*admin)(nil)
var _ StreamMaster = (*master)(nil)

func WithTTL(d time.Duration) ports.CacheOption {
	return func(o *ports.CacheOptions) {
		o.TTL = d
	}
}

func WithNX() ports.CacheOption {
	return func(o *ports.CacheOptions) {
		o.NX = true
	}
}

func WithKeepTTL() ports.CacheOption {
	return func(o *ports.CacheOptions) {
		o.KeepTTL = true
	}
}
