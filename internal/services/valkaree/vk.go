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
	"github.com/zrp9/launchl/internal/config"
	"github.com/zrp9/launchl/internal/crane"
)

type Cacher interface {
	Set(ctx context.Context, key, val string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type Field struct {
	Name  string
	Value string
}

type Message struct {
	ID     string
	Values []Field
}

type Job struct {
	MessageID  string          `json:"messageId"`
	JID        string          `json:"id"`
	Kind       string          `json:"kind"`
	Target     string          `json:"target"`
	Source     string          `json:"source"`
	RetryLimit int64           `json:"retryLimit"`
	Payload    json.RawMessage `json:"payload"`
}

type JobResult struct {
	JID        string
	MsgID      string
	Success    bool
	Error      string
	RetryLimit int64
	Attempts   int64
	Duration   time.Duration
}

type PendingInfo struct {
	Total    int64
	FirstID  string
	LatestID string
	Group    string
}

type StreamEntry map[string]string

type Stream struct {
	client    valkey.Client
	Key       string
	threshold int64
	log       crane.Zlogrus
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
	Count    int64
}

type admin struct {
	s        *Stream
	Group    string
	Consumer string
}

type master struct {
	writer
	reader
	admin
}

type StreamWriter interface {
	Writer(ctx context.Context, fields StreamEntry) (string, error)
	WriteJob(ctx context.Context, kind, target, src string, payload json.RawMessage) (string, error)
	WriteJSON(ctx context.Context, fieldName string, payload []byte) (string, error)
}

type StreamReader interface {
	ReadGroup(ctx context.Context, count int64) ([]Message, error)
	Range(ctx context.Context, start, stop string) ([]Message, error)
	RangeAll(ctx context.Context) ([]Message, error)
	Ack(ctx context.Context, ids ...string) (int64, error)
	AckDel(ctx context.Context, ids ...string) (int64, error)
	ClaimIdle(ctx context.Context, minIdle time.Duration) (string, []Message, error)
	Pending(ctx context.Context) ([]Message, error)
}

type StreamAdmin interface {
	CreateGroup(ctx context.Context) error
	TrimApprox(ctx context.Context) (int64, error)
	DeleteStream(ctx context.Context) error
	DeleteGroup(ctx context.Context) error
	DeleteConsumer(ctx context.Context) error
	StreamInfo(ctx context.Context) error
	GroupInfo(ctx context.Context) error
	ConsumerInfo(ctx context.Context) error
}

type StreamMaster interface {
	StreamWriter
	StreamReader
	StreamAdmin
}

type ValkeyService struct {
	client valkey.Client
}

func NewValkeyService(ctx context.Context) (ValkeyService, error) {
	confi := config.LoadValkey()
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%s", confi.Host, confi.Port)},
	})

	if err != nil {
		return ValkeyService{}, nil
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pong, err := client.Do(pingCtx, client.B().Ping().Build()).ToString()
	if err != nil {
		client.Close()
		return ValkeyService{}, err
	}

	fmt.Println("valkey connected: ", pong)

	return ValkeyService{
		client: client,
	}, nil
}

func (v ValkeyService) Set(ctx context.Context, key, value string) error {
	return v.client.Do(ctx, v.client.B().Set().Key(key).Value(value).Build()).Error()
}

func (v ValkeyService) Get(ctx context.Context, key string) (string, error) {
	return v.client.Do(ctx, v.client.B().Get().Key(key).Build()).ToString()
}

func NewStream(client valkey.Client, key string, threshold int64, log crane.Zlogrus) *Stream {
	return &Stream{
		client:    client,
		Key:       key,
		threshold: threshold,
		log:       log,
	}
}

func (s *Stream) Writer() StreamWriter { return writer{s: s} }

func (s *Stream) Reader(group, consumer string, block time.Duration, count int64) StreamReader {
	return reader{
		s:        s,
		Group:    group,
		Consumer: consumer,
		Block:    block,
		Count:    count,
	}
}

func (s *Stream) Admin(group, consumer string) StreamAdmin {
	return admin{
		s:        s,
		Group:    group,
		Consumer: consumer,
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
			Count:    count,
		},
		admin: admin{
			s:        s,
			Group:    group,
			Consumer: consumer,
		},
	}
}

func (r admin) CreateGroup(ctx context.Context) error {
	res := r.s.client.Do(ctx, r.s.client.B().XgroupCreate().Key(r.s.Key).Group(r.Group).Id("$").Mkstream().Build())

	if err := res.Error(); err != nil {
		var vErr *valkey.ValkeyError
		if errors.As(err, &vErr) && vErr.IsBusyGroup() {
			// group already exists
			return nil
		}
		return err
	}

	return nil
}

func (w writer) Writer(ctx context.Context, fields StreamEntry) (string, error) {
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

func (w writer) WriteJob(ctx context.Context, kind, target, src string, payload json.RawMessage) (string, error) {
	cmd := w.s.client.B().Xadd().Key(w.s.Key).Maxlen().Almost().Threshold(w.s.Threshold()).Id("*").FieldValue().FieldValueIter(func(yield func(string, string) bool) {
		if !yield("jid", uuid.NewString()) {
			return
		}

		if !yield("kind", kind) {
			return
		}

		if !yield("target", target) {
			return
		}

		if !yield("source", src) {
			return
		}

		if !yield("retryLimit", "3") {
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

func (r reader) ReadGroup(ctx context.Context, count int64) ([]Message, error) {
	var c int64
	if count == 0 {
		c = r.Count
	}

	cmd := r.s.client.B().Xreadgroup().Group(r.Group, r.Consumer).Count(c).Block(r.Block.Milliseconds()).Streams().Key(r.s.Key).Id(">").Build()
	res := r.s.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}

	msgs, err := r.s.parseXRead(res)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (r reader) Range(ctx context.Context, start, stop string) ([]Message, error) {
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

func (r reader) RangeAll(ctx context.Context) ([]Message, error) {
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

func (r reader) ClaimIdle(ctx context.Context, minIdle time.Duration) (next string, msgs []Message, err error) {
	cmd := r.s.client.B().Xautoclaim().Key(r.s.Key).Group(r.Group).Consumer(r.Consumer).MinIdleTime(r.s.toMs(minIdle)).Start("0-0").Count(r.Count).Build()

	res := r.s.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		return "", nil, err
	}

	return r.s.parseXAutoClaim(res)
}

func (r reader) Pending(ctx context.Context) ([]Message, error) {
	cmd := r.s.client.B().Xreadgroup().Group(r.Group, r.Consumer).Streams().Key(r.s.Key).Id("0").Build()
	res := r.s.client.Do(ctx, cmd)
	if err := res.Error(); err != nil {
		return nil, err
	}
	return r.s.parseXRead(res)
}

func (r admin) TrimApprox(ctx context.Context) (int64, error) {
	cmd := r.s.client.B().Xtrim().Key(r.s.Key).Maxlen().Almost().Threshold(r.s.Threshold()).Acked().Build()
	return r.s.client.Do(ctx, cmd).ToInt64()
}

func (r admin) DeleteStream(ctx context.Context) error {
	cmd := r.s.client.B().Del().Key(r.s.Key).Build()
	res := r.s.client.Do(ctx, cmd)
	return res.Error()
}

func (r admin) DeleteGroup(ctx context.Context) error {
	cmd := r.s.client.B().XgroupDestroy().Key(r.s.Key).Group(r.Group).Build()
	if err := r.s.client.Do(ctx, cmd).Error(); err != nil {
		r.s.log.MustDebug(fmt.Sprintf("error occurred while deleting stream group %v", err))
		return err
	}
	return nil
}

func (r admin) DeleteConsumer(ctx context.Context) error {
	cmd := r.s.client.B().XgroupDelconsumer().Key(r.s.Key).Group(r.Group).Consumername(r.Consumer).Build()
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

func (r admin) ConsumerInfo(ctx context.Context) error {
	cmd := r.s.client.B().XinfoConsumers().Key(r.s.Key).Group(r.Group).Build()
	res := r.s.client.Do(ctx, cmd)
	log.Printf("x info consumer result %#v", res)
	return nil
}

// Parses Xread / XreadGroup -> map[string][]XRangeSlices where keys are stream names
func (s *Stream) parseXRead(res valkey.ValkeyResult) ([]Message, error) {
	msgs := make([]Message, 0)
	streams, err := res.AsXReadSlices()
	if err != nil {
		return nil, err
	}

	// stream, entries can loop over multiple streams
	for _, entries := range streams {
		for _, e := range entries {
			msgs = append(msgs, Message{
				ID:     e.ID,
				Values: s.parseFields(e.FieldValues),
			})
		}
	}

	return msgs, nil
}

// XRANGE / XREVRANGE → []XRangeSlice
func (s *Stream) parseXRange(res valkey.ValkeyResult) ([]Message, error) {
	slices, err := res.AsXRangeSlices()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]Message, 0, len(slices))
	for _, e := range slices {
		out = append(out, Message{ID: e.ID, Values: s.parseFields(e.FieldValues)})
	}
	return out, nil
}

// XAUTOCLAIM → [ nextStartId, [ entries... ] , (opt: deleted count) ]
func (s *Stream) parseXAutoClaim(res valkey.ValkeyResult) (next string, msgs []Message, err error) {
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
	msgs = make([]Message, 0, len(entrySlices))
	for _, e := range entrySlices {
		msgs = append(msgs, Message{ID: e.ID, Values: s.parseFields(e.FieldValues)})
	}
	return next, msgs, nil
}

func (s *Stream) parseFields(fieldValues []valkey.XRangeFieldValue) []Field {
	fields := make([]Field, 0, len(fieldValues))
	for _, fv := range fieldValues {
		fields = append(fields, Field{
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
