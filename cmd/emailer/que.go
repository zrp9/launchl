// Package emailer reads email events from stream and sends emails
package emailer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/noti"
	"github.com/zrp9/launchl/internal/config"
)

func Main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	vconf := config.LoadValkey()
	logger := crane.DefaultLogger
	if err := run(ctx, vconf, *logger); err != nil {
		log.Printf("failed to load valkeree config %v", err)
	}
}

// TODO: create a env struct for streams
func run(ctx context.Context, vconf config.ValkeyCfg, logger crane.Zlogrus) error {
	valkeyClient, err := valkaree.NewClient(ctx, vconf)
	if err != nil {
		return err
	}

	stream := valkaree.NewStream(valkeyClient, "email-notifications", 1000, logger)
	render, err := noti.NewRenderer()
	if err != nil {
		return err
	}
	// TODO: check these values for reader and make env
	consumer := noti.NewNotiConsumer(stream.Reader("email-notifications", "runner", 500, 1000), render)
	return consumer.Run(ctx)
}
