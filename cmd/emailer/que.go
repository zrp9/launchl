// Package emailer reads email events from stream and sends emails
package emailer

import (
	"context"
	"log"

	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/log/crane"
	"github.com/zrp9/launchl/internal/adapter/noti"
	"github.com/zrp9/launchl/internal/adapter/render"
	"github.com/zrp9/launchl/internal/config"
)

func Main() {
	ctx := context.TODO()
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
	render
	render := render.New()
	consumer := noti.NewNotiConsumer(stream.Reader())

}
