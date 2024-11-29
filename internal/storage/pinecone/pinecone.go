/*
Copyright © 2024 Ben Gittins
*/

package pinecone

import (
	"github.com/henomis/lingoose/index"
	"github.com/henomis/lingoose/index/vectordb/pinecone"
)

type PineCone struct {
	Index index.VectorDB
}

func (idx *PineCone) CreateIndex(idxName string) index.VectorDB {
	idx.Index = pinecone.New(pinecone.Options{
		IndexName: idxName,
		Namespace: "copacetic",
		CreateIndexOptions: &pinecone.CreateIndexOptions{
			Dimension:  0,
			Metric:     "",
			Serverless: nil,
			Pod:        nil,
		},
	})

	return idx.Index
}

func (idx *PineCone) GetDB() index.VectorDB {
	return idx.Index
}
