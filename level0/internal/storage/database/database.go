package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"level0/internal/server"
	"log"
)

type Database struct {
	Conn *sql.DB
}

func (db Database) InsertNewOrder(order server.Order) error {
	rawData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("unable to marshal data: %w", err)
	}

	_, err = db.Conn.Exec("INSERT INTO ordernew2 (order_uid, order_json) VALUES ($1, $2)",
		order.OrderUid, rawData)
	if err != nil {
		return fmt.Errorf("unable to insert data to db: %w", err)
	}

	return nil
}

func (db Database) GetAllOrder() ([]server.Order, error) {
	rows, err := db.Conn.Query("SELECT order_uid, order_json FROM ordernew2")
	if err != nil {
		log.Println("Query select from db ERROR.", err)
	}

	orders := make([]server.Order, 0)

	//ADDING DATA INTO STRUCT WITH STRING TYPE FIELDS
	for rows.Next() {
		var (
			id      string
			rawJson string
		)

		if err := rows.Scan(&id, &rawJson); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		var order server.Order

		err := json.Unmarshal([]byte(rawJson), &order)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal order data: %w", err)
		}

		order.OrderUid = id

		orders = append(orders, order)
	}

	return orders, nil
}

func (db Database) GetOrderByID(orderID string) (*server.Order, error) {
	var (
		id      string
		rawJson string
	)

	err := db.Conn.QueryRow("SELECT order_uid, order_json FROM ordernew2 WHERE order_uid = $1",
		orderID).Scan(&id, &rawJson)
	if err != nil {
		return nil, err
	}

	var order server.Order

	err = json.Unmarshal([]byte(rawJson), &order)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal order data: %w", err)
	}

	order.OrderUid = id

	return &order, nil
}

// NewDatabaseConnect CONNECTING TO DATABASE
func NewDatabaseConnect(dbConfig Config) (*Database, error) {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open sql ERROR: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db ERROR: %w", err)
	}

	return &Database{Conn: db}, nil
}
