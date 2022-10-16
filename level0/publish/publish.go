package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type Order struct {
	OrderUid    string `json:"order_uid"`
	TrackNumber string `json:"track_number"`
	Entry       string `json:"entry"`
	Delivery    struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
		Transaction  string `json:"transaction"`
		RequestId    string `json:"request_id"`
		Currency     string `json:"currency"`
		Provider     string `json:"provider"`
		Amount       int    `json:"amount"`
		PaymentDt    int    `json:"payment_dt"`
		Bank         string `json:"bank"`
		DeliveryCost int    `json:"delivery_cost"`
		GoodsTotal   int    `json:"goods_total"`
		CustomFee    int    `json:"custom_fee"`
	} `json:"payment"`
	Items             []Items   `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerId        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmId              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Items struct {
	ChrtId      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmId        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func main() {
	//MODEL.JSON TO BYTES
	btsJsn, err := ioutil.ReadFile("./publish/model.json")
	if err != nil {
		log.Println("ReadFile model.json ERROR.", err)
	}

	//CONNECTING TO NATS-STREAMING-SERVER
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Println("nats.Connect from publisher ERROR.", err)
	}
	log.Println(nc.IsConnected())

	//ENCODED CONNECTION
	sc, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Println("EncodedConnect from main ERROR.", err)
	}
	//log.Println("EncodedConnection.")
	defer sc.Close()

	var order Order

	err = json.Unmarshal(btsJsn, &order)
	if err != nil {
		log.Println("Unmarshal btsJsn to order ERROR.", err)
	}

	//PUBLISHING MODEL.JSON
	if err := sc.Publish("order", order); err != nil {
		log.Println("sc.Publish ERROR.", err)
	}

	sc.Flush()
}
