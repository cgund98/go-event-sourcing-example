package controller

import (
	"github.com/cgund98/go-eventsrc-example/internal/entity/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
)

type Controller struct {
	store    eventsrc.Store
	producer eventsrc.Producer

	transactor     pg.Transactor
	projectionRepo orders.ProjectionRepo
}

func NewController(store eventsrc.Store, producer eventsrc.Producer, projectionRepo orders.ProjectionRepo, transactor pg.Transactor) *Controller {
	return &Controller{store: store, producer: producer, projectionRepo: projectionRepo, transactor: transactor}
}
