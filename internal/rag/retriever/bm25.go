package retriever

import (
	"database/sql"
	"fmt"
)

type BM25Retriever struct {
	db *sql.DB
	
}

func NewBM25Retriever() *BM25Retriever {
	fmt.Println("BM25 Retriever initialized")
	return &BM25Retriever{}
}
