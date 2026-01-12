package retriever

import "testing"

func TestBM25Search(t *testing.T) {
	retriever := NewBM25Retriever()
	if retriever == nil {
		t.Error("Expected BM25Retriever instance, got nil")
	}
}
