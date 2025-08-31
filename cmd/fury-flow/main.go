package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/petervincek/fury-flow/internal/db"
)

func getKanbanBoards() {
	dbConn, err := sql.Open("pgx", "postgres://admin:password@localhost:5432/furyflow?sslmode=disable")
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
	getKanbanBoards()
}
