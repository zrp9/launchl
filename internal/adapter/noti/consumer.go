// Package noti contains email consumers
package noti

import (
	"github.com/zrp9/launchl/internal/adapter/cache/valkaree"
	"github.com/zrp9/launchl/internal/adapter/render"
)

type NotiConsumer struct {
	streamReader  valkaree.StreamReader
	emailRenderer render.Renderer
}

func NewNotiConsumer(sreader valkaree.StreamReader, erenderer render.Renderer) NotiConsumer {
	return NotiConsumer{
		streamReader:  sreader,
		emailRenderer: erenderer,
	}
}
