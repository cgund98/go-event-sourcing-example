package controller

import (
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
)

type Controller struct {
	store    eventsrc.Store
	producer eventsrc.Producer
}

func NewController(store eventsrc.Store, producer eventsrc.Producer) *Controller {
	return &Controller{store: store, producer: producer}
}
