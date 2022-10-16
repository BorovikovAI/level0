package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"level0/internal/server"
	"level0/internal/storage/broker"
	"level0/internal/storage/database"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"

	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
)

type OrderIsReady struct {
	OrderUid  string    `json:"order_uid"`
	OrderNoID OrderNoID `json:"order_json"`
}

type OrderIsReadyScan struct {
	OrderUid  string `json:"order_uid"`
	OrderJson string `json:"order_json"`
}

type OrderNoID struct {
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

type ViewData struct {
	OrderId   string
	OrderInfo string
}

type Cache struct {
	OrderCache *cache.Cache
}

func readConfig(cfgPath string) (*Config, error) {
	cfg := &Config{}

	bytesCfg, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read cfg file: %w", err)
	}

	err = json.Unmarshal(bytesCfg, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to convert cfg to model: %w", err)
	}

	return cfg, nil
}

func main() {
	cfg, err := readConfig("./cmd/level0/config.json")
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewDatabaseConnect(cfg.DB)
	if err != nil {
		log.Fatal(err)
	}

	natsBroker, err := broker.NewNatsConnect(cfg.Nats)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(db, natsBroker)
	if err != nil {
		log.Fatal(err)
	}

	s.Run()
}

func main2() {
	var order Order

	//CONNECTING TO NATS-STREAMING-SERVER
	nc, err := nats.Connect("nats://127.0.0.1:4222")
	if err != nil {
		log.Println("nats.Connect from main ERROR.", err)
	}
	log.Println("Connection:", nc.IsConnected())

	//ENCODED CONNECTION
	sc, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Println("EncodedConnect from main ERROR.", err)
	}
	log.Println("EncodedConnection:", sc.Conn.Statistics)
	defer sc.Close()

	//BASIC SUBSCRIPTION
	/*sub, err := sc.SubscribeSync("order")
	if err != nil {
		log.Println("sc.SubscribeSync ERROR.", err)
	}
	log.Println("Subscribed.")*/

	//ENCODED SUBSCRIPTION
	sub, err := sc.Subscribe("order", func(order *Order) {
		fmt.Printf("Received an order: %+v\n", order)
	})
	if err != nil {
		log.Println("EncodedSubscription ERROR.", err)
	}
	log.Println("EncodedSubscription.")

	recvCh := make(chan map[string]interface{})
	orderMap := make(map[string]interface{})

	sc.BindRecvChan("order", recvCh)

	orderMap = <-recvCh

	orderJsonData, err := json.Marshal(orderMap)
	if err != nil {
		log.Println("Marshal encoded ERROR.", err)
	}

	err = json.Unmarshal(orderJsonData, &order)
	if err != nil {
		log.Println("Unmarshal encoded ERROR.", err)
	}
	log.Println(order)

	/*m, err := sub.NextMsg(time.Hour)
	if err != nil {
		log.Println("sub.NextMsg ERROR.", err)
	}*/

	sub.Unsubscribe()
	log.Println("Unsubscribed.")

	//PARSING RECIEVED DATA INTO STRUCT "order"
	/*err = json.Unmarshal(m.Data, &order)
	if err != nil {
		log.Println("Unmarshal ERROR.", err)
	}
	log.Println(order)*/

	//SLICING "OrderUid" FROM STRUCT "order"
	var orderIsReady OrderIsReady

	orderIsReady.OrderUid = order.OrderUid

	var orderNoID OrderNoID

	orderNoID.TrackNumber = order.TrackNumber
	orderNoID.Entry = order.Entry

	orderNoID.Delivery = order.Delivery
	orderNoID.Payment = order.Payment
	orderNoID.Items = order.Items

	orderNoID.Locale = order.Locale
	orderNoID.InternalSignature = order.InternalSignature
	orderNoID.CustomerId = order.CustomerId
	orderNoID.DeliveryService = order.DeliveryService
	orderNoID.Shardkey = order.Shardkey
	orderNoID.SmId = order.SmId
	orderNoID.DateCreated = order.DateCreated
	orderNoID.OofShard = order.OofShard

	orderIsReady.OrderNoID = orderNoID

	//CREATING NEW CACHE
	var orderCache Cache

	orderCache.OrderCache = cache.New(cache.DefaultExpiration, cache.DefaultExpiration)

	//PARSING DATA FROM STRUCT "orderIsReady" TO BYTES
	orderNoIDMarshaled, err := json.Marshal(orderIsReady.OrderNoID)
	if err != nil {
		log.Println("Marshal ERROR.", err)
	}
	log.Println(orderNoIDMarshaled)

	//ADDING ORDER DATA INTO THE CACHE
	err = orderCache.OrderCache.Add(fmt.Sprintf("%+v", orderIsReady.OrderUid), string(orderNoIDMarshaled), cache.DefaultExpiration)
	if err != nil {
		log.Println("Cache Add ERROR.", err)
	}

	//SAVING CACHE
	err = orderCache.OrderCache.SaveFile("cache")
	if err != nil {
		log.Println("SaveFile cache ERROR.", err)
	}

	//CONNECTING TO DATABASE
	db, err := sql.Open("postgres", "user=postgres password=pass1488 dbname=testdb sslmode=disable")
	if err != nil {
		log.Println("Open sql ERROR.", err)
	}

	if err := db.Ping(); err != nil {
		log.Println("Ping db ERROR.", err)
	}

	//INSERTING DATA INTO POSTGRESQL
	/*query := "INSERT INTO ordernew2 (order_uid, order_json) VALUES ($1, $2)"
	result, err := db.Exec(query, orderIsReady.OrderUid, orderNoIDMarshaled)
	if err != nil {
		log.Println("Exec db ERROR.", err)
	}
	log.Println(result)*/

	var orderIsReadyScan OrderIsReadyScan

	//SELECTING DATA FROM POSTGRESQL
	querySelect := "SELECT * FROM ordernew WHERE order_uid=$1"
	rows, err := db.Query(querySelect, orderIsReady.OrderUid)
	if err != nil {
		log.Println("Query select from db ERROR.", err)
	}

	//ADDING DATA INTO STRUCT WITH STRING TYPE FIELDS
	for rows.Next() {
		if err := rows.Scan(&orderIsReadyScan.OrderUid, &orderIsReadyScan.OrderJson); err != nil {
			log.Println("Scan rows ERROR.", err)
		}
		/*err := orderCache.OrderCache.Add(fmt.Sprintf("%+v", orderIsReadyScan.OrderUid), string(orderIsReadyScan.OrderJson), cache.DefaultExpiration)
		if err != nil {
			log.Println("Cache Add ERROR.", err)
		}*/
	}

	/*err = orderCache.OrderCache.SaveFile("cache")
	if err != nil {
		log.Println("SaveFile cache ERROR.", err)
	}*/

	orderIsReadyScan.OrderJson = strings.Replace(orderIsReadyScan.OrderJson, "\"", "", -1)
	log.Println(orderIsReadyScan.OrderUid, orderIsReadyScan.OrderJson)

	/*http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(orderIsReadyScan)
		if err != nil {
			log.Println("Marshal orderIsReady to bytes ERROR.", err)
		}
		w.Write(bytes)
	})*/

	//HTTP SERVER
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handled /")
		http.ServeFile(w, r, "htmls/index.html")
	})

	http.HandleFunc("/postform", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handled /postform")

		orderId := r.FormValue("orderId")

		iOrderJson, flag := orderCache.OrderCache.Get(orderId)

		var data ViewData

		if flag {
			var strOrderJson string = fmt.Sprintf("%+v", iOrderJson)
			data = ViewData{
				OrderId:   orderId,
				OrderInfo: strOrderJson,
			}
		} else {
			data = ViewData{
				OrderId:   "NO DATA: (OrderId).",
				OrderInfo: "NO DATA: (OrderInfo).",
			}
		}

		tmpl, err := template.ParseFiles("htmls/order_id.html")
		if err != nil {
			log.Println("ParseFiles template into tmpl ERROR.", err)
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println("Execute tmpl ERROR.", err)
		}
	})

	http.ListenAndServe("localhost:8003", nil)
}
