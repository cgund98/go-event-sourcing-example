package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	pb "github.com/cgund98/go-eventsrc-example/api/v1/orders"
	"github.com/cgund98/go-eventsrc-example/internal/infra/config"
	"github.com/cgund98/go-eventsrc-example/internal/infra/eventsrc"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"github.com/cgund98/go-eventsrc-example/internal/infra/pg"
	"github.com/cgund98/go-eventsrc-example/internal/service/orders"

	"buf.build/go/protovalidate"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

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

func initKafkaWriter(config *config.Config) (*kafka.Writer, func(), error) {
	kafkaConnStr := fmt.Sprintf("%s:%d", config.KafkaHost, config.KafkaPort)
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaConnStr},
		Topic:   config.EventsTopic,
	})

	cleanup := func() {
		if err := writer.Close(); err != nil {
			logging.Logger.Error(fmt.Sprintf("error closing kafka connection: %v", err))
		}
	}

	return writer, cleanup, nil
}

func runGRPCServer(config *config.Config, producer *eventsrc.TransactionProducer) error {
	orderService := orders.NewOrderService(producer)

	// Create a Protovalidate Validator
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("unable to create protovalidate validator: %v", err)
	}

	// Use the protovalidate_middleware interceptor provided by grpc-ecosystem
	interceptor := protovalidate_middleware.UnaryServerInterceptor(validator)

	server := grpc.NewServer(grpc.UnaryInterceptor(interceptor))
	pb.RegisterOrderServiceServer(server, orderService)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GrpcPort))
	if err != nil {
		return fmt.Errorf("unable to listen on gRPC port: %v", err)
	}
	defer lis.Close()

	logging.Logger.Info("Starting gRPC server...", "address", lis.Addr().String())

	// Start gRPC server
	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %v", err)
	}

	return nil
}

func runGatewayServer(ctx context.Context, config *config.Config) error {
	// Create gRPC-Gateway mux with JSON marshaler that uses snake_case
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{},
		}),
	)

	// gRPC server address for gateway to connect to
	grpcAddr := fmt.Sprintf("localhost:%d", config.GrpcPort)

	// Register gRPC-Gateway handler
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterOrderServiceHandlerFromEndpoint(ctx, gwmux, grpcAddr, opts)
	if err != nil {
		return fmt.Errorf("failed to register gateway handler: %v", err)
	}

	// Start HTTP server
	gatewayAddr := fmt.Sprintf(":%d", config.HttpPort)
	logging.Logger.Info("Starting gRPC-Gateway server...", "address", gatewayAddr)

	// Create HTTP server with context
	server := &http.Server{
		Addr:    gatewayAddr,
		Handler: gwmux,
	}

	// Start gateway server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve gateway: %v", err)
	}

	return nil
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
	kafkaWriter, cleanup, err := initKafkaWriter(config)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("unable to initialize kafka: %v", err))
		os.Exit(1)
	}
	defer cleanup()

	logging.Logger.Info("Starting order service...")

	store := eventsrc.NewPostgresStore(db, config.EventsTable)
	bus := eventsrc.NewKafkaBus(kafkaWriter)
	tx := pg.NewDbTransactor(db)

	producer := eventsrc.NewTransactionProducer(store, bus, tx, config.EventsTopic)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create error group for managing servers
	g, ctx := errgroup.WithContext(ctx)

	// Start gRPC server
	g.Go(func() error {
		return runGRPCServer(config, producer)
	})

	// Start gRPC-Gateway server
	g.Go(func() error {
		return runGatewayServer(ctx, config)
	})

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		logging.Logger.Error(fmt.Sprintf("server error: %v", err))
		os.Exit(1)
	}

	logging.Logger.Info("All servers shut down gracefully")
}
