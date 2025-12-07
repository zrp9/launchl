package main

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

func main() {
	log.Println("email worker running...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	cfg := config.LoadVkENotifierConfig()

	logger := crane.DefaultLogger
	if err := run(ctx, cfg, *logger); err != nil {
		log.Printf("failed to load valkeree config %v", err)
	}
}

// TODO: create a env struct for streams
func run(ctx context.Context, conf config.VkNotifierCfg, logger crane.Zlogrus) error {
	valkeyClient, err := valkaree.NewClient(ctx, conf.ClientCfg)
	if err != nil {
		return err
	}

	stream := valkaree.NewStream(valkeyClient, conf.StreamCfg.Name, conf.StreamCfg.WriteRetries, conf.StreamCfg.Count, logger)
	admin := stream.Admin()
	if err = admin.CreateGroup(ctx, conf.StreamCfg.Group); err != nil {
		log.Printf("failed to create group %v", err)
		return err
	}

	render, err := noti.NewRenderer()
	if err != nil {
		return err
	}
	notificationProcessor := noti.NewNotifer(
		stream.Reader(
			conf.StreamCfg.Group,
			conf.StreamCfg.Consumer,
			conf.StreamCfg.Block,
			conf.StreamCfg.Count,
		),
		noti.NewSender(
			conf.SenderCfg.Host,
			conf.SenderCfg.Sender,
			conf.SenderCfg.Username,
			conf.SenderCfg.Token,
			conf.SenderCfg.Port,
			*crane.DefaultLogger,
		),
		render,
	)
	return notificationProcessor.Run(ctx)
}
