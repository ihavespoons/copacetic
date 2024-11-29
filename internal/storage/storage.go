/*
Copyright © 2024 Ben Gittins
*/

package storage

import "github.com/henomis/lingoose/index"

type Storage interface {
	CreateIndex(idxName string) index.VectorDB
	GetIndex() index.VectorDB
}
