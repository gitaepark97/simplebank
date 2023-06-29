package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/gitaepark/simplebank/api"
	db "github.com/gitaepark/simplebank/db/sqlc"
	_ "github.com/gitaepark/simplebank/doc/statik"
	"github.com/gitaepark/simplebank/gapi"
	"github.com/gitaepark/simplebank/pb"
	"github.com/gitaepark/simplebank/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func runDBMigration(migrationURL, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal("cannot create new migrate instace", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up:", err)
	}

	log.Println("db migrated successfully")
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err =  grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server")
	}
}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
 	if err != nil {
 		log.Fatal("cannot create server:", err)
 	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

 	grpcMux := runtime.NewServeMux(jsonOption)

 	ctx, cancel := context.WithCancel(context.Background())
 	defer cancel()

 	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
 	if err != nil {
 		log.Fatal("cannot register handler server:", err)
 	}

 	mux := http.NewServeMux()
 	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs:", err)
	}
	swaggerHandelr := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandelr)

 	listener, err := net.Listen("tcp", config.HTTPServerAddress)
 	if err != nil {
 		log.Fatal("cannot create listener:", err)
 	}

 	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
 	err = http.Serve(listener, mux)
 	if err != nil {
 		log.Fatal("cannot start HTTP gateway server:", err)
 	}
}

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)
	// runGinServer(config, store)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)
	
}