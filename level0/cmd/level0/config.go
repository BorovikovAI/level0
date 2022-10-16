package main

import (
	"level0/internal/storage/broker"
	"level0/internal/storage/database"
)

type Config struct {
	DB   database.Config `json:"DB"`
	Nats broker.Config   `json:"nats"`
}
