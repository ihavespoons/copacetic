/*
Copyright © 2024 Ben Gittins
*/

package llm

import (
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/rag"
)

type LLM interface {
	CreateIndex(sugarAddr string) error
	CreateRAG(directory string) error
	ModelType() string
}

type Model struct {
	Temperature      float64
	Index            *index.Index
	DocumentationRAG *rag.RAG
	SourceRAG        *rag.RAG
}
