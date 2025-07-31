package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cgund98/go-eventsrc-example/internal/infra/config"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/segmentio/kafka-go"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const EventsTable = "events"

func initDB(config *config.Config) (*sqlx.DB, func(), error) {
	// Start DB Connection
	dbConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		config.PostgresHost, config.PostgresPort, config.PostgresUser, config.PostgresPassword, config.PostgresDB,
	)
	if config.PostgresHost == "localhost" {
		dbConnStr = fmt.Sprintf("%s sslmode=disable", dbConnStr)
	}

	db, err := sqlx.Connect("postgres", dbConnStr)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("unable to initialize db: %v", err))
		os.Exit(1)
	}
	cleanup := func() {
		if err := db.Close(); err != nil {
			logging.Logger.Error(fmt.Sprintf("error closing db connection: %v", err))
		}
	}

	return db, cleanup, nil
}

func initKafka(config *config.Config) (*kafka.Conn, func(), error) {
	kafkaConnStr := fmt.Sprintf("%s:%d", config.KafkaHost, config.KafkaPort)
	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaConnStr, config.EventsTopic, 0)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		if err := conn.Close(); err != nil {
			logging.Logger.Error(fmt.Sprintf("error closing kafka connection: %v", err))
		}
	}

	return conn, cleanup, nil
}

func main() {

	// Load Config
	config, err := config.LoadConfig()
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("unable to load config: %v", err))
		os.Exit(1)
	}

	// Initialize DB
	db, cleanup, err := initDB(config)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("unable to initialize db: %v", err))
		os.Exit(1)
	}
	defer cleanup()

	// Initialize Kafka
	kafkaConn, cleanup, err := initKafka(config)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("unable to initialize kafka: %v", err))
		os.Exit(1)
	}
	defer cleanup()

	logging.Logger.Info("Starting order service...")

	store := eventsrc.NewPostgresStore(db, config.EventsTable)
	bus := eventsrc.NewKafkaBus(kafkaConn)
	tx := pg.NewDbTransactor(db)

	producer := eventsrc.NewTransactionProducer(store, bus, tx, config.EventsTopic)

	logging.Logger.Info("Sending event...")
	err = producer.Send(context.Background(), &eventsrc.SendArgs{
		AggregateID: "123",
		EventType:   "OrderCreated",
		Value:       []byte(`{"amount": 100}`),
	})
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("unable to send event: %v", err))
		os.Exit(1)
	}

	logging.Logger.Info("Event sent successfully")

	// Print all events in the bus
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{fmt.Sprintf("%s:%d", config.KafkaHost, config.KafkaPort)},
		Topic:   config.EventsTopic,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logging.Logger.Error(fmt.Sprintf("unable to read message: %v", err))
			os.Exit(1)
		}
		eventTypeHeader := string(msg.Headers[0].Value)
		logging.Logger.Info(fmt.Sprintf("message: %s", string(msg.Value)), "eventTypeHeader", eventTypeHeader)
	}
}
