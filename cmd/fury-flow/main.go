package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/petervincek/fury-flow/internal/config"
	"github.com/petervincek/fury-flow/internal/db"
	"github.com/petervincek/fury-flow/internal/logging"
)

var logger = logging.GetLogger()

func getKanbanBoards() {
	dbDriver := os.Getenv("DB_DRIVER")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Fatal("unable to parse db port", err)
	}
	dbName := os.Getenv("DB_NAME")
	logger.Info("About to connect to PostgreSQL")
	dbConn, err := sql.Open("pgx", fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable", dbDriver, dbUsername, dbPassword, dbHost, dbPort, dbName))

	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	q := db.New(dbConn)
	boards, err := q.GetKanbanBoards(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, board := range boards {
		fmt.Printf("Board: %s, desc: %s\n", board.BoardName, board.Description.String)
	}
}

func main() {
	fmt.Printf("Fury flow\n")
	config.LoadEnvVariables(".env")
	getKanbanBoards()
}
