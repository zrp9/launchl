// Package noti contains email consumers
package noti

import (
	"context"
	"log"
	"time"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/domain"
	"github.com/zrp9/launchl/internal/domain/core"
)

type EmailJob struct {
	core.Email
	// To              []string       `json:"to"`
	// From            string         `json:"from"`
	// Subject         string         `json:"subject,omitempty"`
	// Data            map[string]any `json:"data,omitempty"`
	Template        string `json:"template"`
	TemplateVersion string `json:"templateVersion,omitempty"`
}

type Notifier struct {
	reader        valkaree.StreamReader
	emailRenderer *Renderer
	sender        Sender
	sendRetries   int
	expBackoff    int
}

func NewNotifer(sr valkaree.StreamReader, ns Sender, r *Renderer) Notifier {
	log.Printf("Notifier using StreamReader type: %T\n", sr)
	return Notifier{
		reader:        sr,
		emailRenderer: r,
		sender:        ns,
	}
}

func (c Notifier) Run(ctx context.Context) error {
	log.Printf("Notifier using StreamReader type: %T\n", c.reader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			events, err := c.reader.ReadGroup(ctx, c.reader.Count())
			log.Printf("email worker reading msg %v", events)
			if err != nil {
				log.Printf("error durring read group %v", err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			if len(events) == 0 {
				// xreadgroup timeout - loop again
				continue
			}

			for _, e := range events {
				if err := c.handle(ctx, e); err != nil {
					log.Printf("failed to handle msg %v", err)
					// ignore for now and will be picked back up maybe eventually write to failed stream
					continue
				}

				_, _ = c.reader.AckDel(ctx, e.ID)
			}
		}
	}
}

func (c Notifier) handle(ctx context.Context, msg domain.Message) error {
	job, err := parseEmailJob(msg)
	if err != nil {
		return err
	}

	// eventType, err := parseEventType(msg)
	// log.Printf("event type parsed %v", eventType)
	// if err != nil {
	// 	return err
	// }

	html, text, err := c.emailRenderer.Render(job.Template, job)
	if err != nil {
		return err
	}

	var last error
	for attempt := 0; attempt <= c.sendRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(c.expBackoff) * time.Duration(attempt)):
			}
		}
		if err := c.sender.Send(ctx, job.To, job.Subject, html, text); err != nil {
			last = err
			continue
		}

		return nil
	}
	return last
}
