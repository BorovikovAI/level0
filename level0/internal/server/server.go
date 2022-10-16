package server

import (
	"errors"
	"fmt"
	"html/template"
	"level0/internal/storage/broker"
	"log"
	"net/http"

	"github.com/patrickmn/go-cache"
)

type Database interface {
	InsertNewOrder(order Order) error
	GetAllOrder() ([]Order, error)
	GetOrderByID(orderID string) (*Order, error)
}

type Server struct {
	Broker     *broker.Nats
	DB         Database
	OrderCache *cache.Cache
}

func (s Server) readDataFromBroker() error {
	recvCh := make(chan Order)
	s.Broker.EncodedConn.BindRecvChan("order", recvCh)

	order := <-recvCh
	err := s.DB.InsertNewOrder(order)
	if err != nil {
		log.Println(err)
	}

	err = s.OrderCache.Add(
		order.OrderUid,
		fmt.Sprintf("%+v", order),
		cache.DefaultExpiration,
	)
	if err != nil {
		return fmt.Errorf("cannot add data to cache: %w", err)
	}

	return nil
}

func (s Server) initCache() error {
	orders, err := s.DB.GetAllOrder()
	if err != nil {
		return fmt.Errorf("unable to get order from storage: %w", err)
	}

	for orIdx := range orders {
		err = s.OrderCache.Add(
			orders[orIdx].OrderUid,
			fmt.Sprintf("%+v", orders[orIdx]),
			cache.DefaultExpiration,
		)
		if err != nil {
			return fmt.Errorf("cannot add data to cache: %w", err)
		}
	}

	return nil
}

func (s Server) GerOrderForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./internal/server/htmls/index.html")
}

func (s Server) GerOrderByID(w http.ResponseWriter, r *http.Request) {
	orderId := r.FormValue("orderId")
	if orderId == "" {
		responseErr := errors.New("empty value in orderId")

		log.Print(responseErr)

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(responseErr.Error()))
	}

	iOrderJson, exist := s.OrderCache.Get(orderId)
	if !exist {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	data := ViewData{
		OrderId:   orderId,
		OrderInfo: fmt.Sprintf("%+v", iOrderJson),
	}

	tmpl, err := template.ParseFiles("./internal/server/htmls/order_id.html")
	if err != nil {
		responseErr := fmt.Errorf("parseFiles template into tmpl ERROR: %w", err)

		log.Print(responseErr)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(responseErr.Error()))
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		responseErr := fmt.Errorf("execute tmpl ERROR: %w", err)

		log.Print(responseErr)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(responseErr.Error()))
	}
}

func (s Server) Run() {
	go func() {
		for {
			if err := s.readDataFromBroker(); err != nil {
				log.Println(err)
			}
		}
	}()

	http.ListenAndServe("localhost:8000", nil)
}

func NewServer(database Database, nats *broker.Nats) (*Server, error) {
	server := &Server{
		Broker:     nats,
		DB:         database,
		OrderCache: cache.New(cache.DefaultExpiration, cache.DefaultExpiration),
	}

	err := server.initCache()
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/", server.GerOrderForm)

	http.HandleFunc("/postform", server.GerOrderByID)

	return server, nil
}
