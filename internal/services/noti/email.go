// Package noti is a library for notifications
package noti

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zrp9/launchl/internal/crane"
	vk "github.com/zrp9/launchl/internal/services/valkaree"
)

type Notifier interface {
	Send(ctx context.Context, job vk.Job) error
}

type EmailJob struct {
	To              []string       `json:"to"`
	From            string         `json:"from"`
	Data            map[string]any `json:"data,omitempty"`
	Template        string         `json:"template"`
	TemplateVersion string         `json:"templateVersion,omitempty"`
	Subject         string         `json:"subject,omitempty"`
}

type EmailNoti struct{}

func (e EmailNoti) Send(ctx context.Context, job vk.Job) error {
	payload, err := e.decodeEmailJob(job)
	if err != nil {
		return err
	}

	log.Printf("payload %v", payload)
	return nil
}

type EmailQueConsumer struct {
	streamReader vk.StreamReader
	logger       crane.Zlogrus
	MaxWorkers   int
	Retries      int64
	Timeout      time.Duration
	MinIdle      time.Duration
	Notifier     EmailNoti
}

func NewEmailConsumer(reader vk.StreamReader, emailNoti EmailNoti, retries int64, maxRoutines int, timeout, minIdle time.Duration) EmailQueConsumer {
	return EmailQueConsumer{
		streamReader: reader,
		MaxWorkers:   maxRoutines,
		Retries:      retries,
		Timeout:      timeout,
		MinIdle:      minIdle,
		Notifier:     emailNoti,
	}
}

func (e EmailQueConsumer) Run(ctx context.Context) error {
	jobs := make(chan vk.Message, e.MaxWorkers)
	results := make(chan vk.JobResult, e.MaxWorkers)

	var workers sync.WaitGroup
	for w := 1; w <= e.MaxWorkers; w++ {
		workers.Add(1)
		go e.processMessage(ctx, jobs, results, &workers)
	}

	var resultWg sync.WaitGroup
	go e.monitorResults(ctx, results, &resultWg)

	defer func() {
		close(jobs)
		workers.Wait()
		close(results)
		resultWg.Wait()
	}()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		free := cap(jobs) - len(jobs)
		msgs, err := e.streamReader.ReadGroup(ctx, int64(free))
		if err != nil || len(msgs) == 0 {
			continue
		}

		for _, msg := range msgs {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case jobs <- msg:
			}
		}
	}

}

func (e EmailQueConsumer) processMessage(ctx context.Context, msgs <-chan vk.Message, results chan<- vk.JobResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgs:
			if !ok {
				return
			}
			log.Printf("message %s recieved", m)
			// need to update this to either handle a single field named json or multiple field value pairs
			// extract json payload
			job, err := e.decodeJob(m)
			if err != nil {
				_, _ = e.streamReader.AckDel(ctx, m.ID)
				e.logger.MustDebug(fmt.Sprintf("invalid message will be deleted: msgId: %v, %v", m.ID, err))
			}

			// sendCtx := ctx
			// if e.Timeout > 0 {
			// 	var cancel context.CancelFunc
			// 	sendCtx, cancel = context.WithTimeout(ctx, e.Timeout)
			// 	defer func() {
			// 		cancel()
			// 	}()
			// }
			if err := e.Notifier.Send(ctx, job); err != nil {
				results <- vk.JobResult{Success: false, Error: err.Error(), MsgID: m.ID}
			}

			if _, err := e.streamReader.AckDel(ctx, m.ID); err != nil {
				results <- vk.JobResult{Success: false, Error: fmt.Sprintf("ackdel failed: %v", err), MsgID: m.ID}
			}

			results <- vk.JobResult{Success: true, MsgID: m.ID}

		}
	}
}

// todo handle single json field
func (e EmailQueConsumer) decodeJob(msg vk.Message) (vk.Job, error) {
	values := make(map[string]string, len(msg.Values))
	for _, field := range msg.Values {
		values[field.Name] = field.Value
	}

	payload, ok := values["payload"]
	if !ok {
		payload, ok = values["json"]
		if !ok {
			return vk.Job{}, errors.New("message is missing a payload")
		}
	}

	jid, ok := values["jid"]
	if !ok {
		return vk.Job{}, errors.New("message is missing job id")
	}

	job := vk.Job{MessageID: msg.ID, JID: jid, Payload: json.RawMessage(payload)}
	if kind, ok := values["kind"]; ok {
		job.Kind = kind
	}

	if target, ok := values["target"]; ok {
		job.Target = target
	}

	if src, ok := values["source"]; ok {
		job.Source = src
	}

	if retries, ok := values["retryLimit"]; ok {
		r, err := e.toInt64(retries)
		if err != nil {
			return vk.Job{}, err
		}
		job.RetryLimit = r
	}

	return job, nil
}

func (e EmailNoti) decodeEmailJob(job vk.Job) (EmailJob, error) {
	var ejob EmailJob
	if err := json.Unmarshal([]byte(job.Payload), &ejob); err != nil {
		return EmailJob{}, err
	}

	return ejob, nil
}

func (e EmailNoti) splitCsv(values string) []string {
	return strings.Split(values, ",")
}

func (e EmailQueConsumer) toInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (e EmailQueConsumer) monitorResults(ctx context.Context, results <-chan vk.JobResult, wg *sync.WaitGroup) {
	defer wg.Done()
	var ok, fail int
	for {
		select {
		case <-ctx.Done():
			return
		case r, okCh := <-results:
			if okCh {
				e.logger.MustTrace(fmt.Sprintf("email results: success=%d fail=%d", ok, fail))
			}
			if r.Success {
				ok++
			} else {
				fail++
			}
		}
	}
}
