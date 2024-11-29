/*
Copyright © 2024 Ben Gittins
*/

package openai

import (
	"context"
	"copacetic/internal/llm"
	"copacetic/internal/repository"
	"errors"
	"github.com/henomis/lingoose/document"
	openaiembedder "github.com/henomis/lingoose/embedder/openai"
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/pinecone"
	"github.com/henomis/lingoose/rag"
	"github.com/henomis/lingoose/types"
	"github.com/sourcegraph/conc/pool"
	"os"
)

type OpenAI struct {
	model llm.Model
}

func (oAI *OpenAI) CreateIndex(repositoryName string) error {
	oAI.model.Index = index.New(, openaiembedder.New(openaiembedder.AdaEmbeddingV2)).WithIncludeContents(true)
	if oAI.model.Index == nil {
		return errors.New("index failed to setup correctly")
	}
	return nil
}

func ragWorker(ctx context.Context, rag *rag.RAG, file *repository.File) error {
	fileContents, _ := os.ReadFile(file.Path)
	return rag.AddDocuments(ctx, document.Document{Content: string(fileContents), Metadata: types.Meta{"type": "source"}})
}

func (oAI *OpenAI) CreateRAG(repository *repository.Source) error {
	oAI.model.SourceRAG = rag.New(
		oAI.model.Index,
	).WithChunkSize(1000).WithChunkOverlap(0)
	oAI.model.DocumentationRAG = rag.New(
		oAI.model.Index,
	).WithChunkSize(1000).WithChunkOverlap(0)

	// Process the source code first
	p := pool.New().WithContext(context.Background())

	for _, file := range repository.Source {
		file := file
		p.Go(func(ctx context.Context) error {
			return ragWorker(ctx, oAI.model.SourceRAG, file)
		})
	}
	err := p.Wait()

	return err
}

func (oAI *OpenAI) ModelType() string {
	return "openai"
}

func New(temperature float64, repository *repository.Source) (*OpenAI, error) {
	llmEngine := &OpenAI{
		llm.Model{Temperature: temperature},
	}

	err := llmEngine.CreateIndex()

	if err != nil {
		return nil, err
	}

	err = llmEngine.CreateRAG(repository)

	if err != nil {
		return nil, err
	}

	return llmEngine, nil
}
