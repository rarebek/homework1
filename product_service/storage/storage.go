package storage

import (
	"EXAM3/product_service/pkg/db"
	"EXAM3/product_service/pkg/logger"
	"EXAM3/product_service/storage/postgres"
	"EXAM3/product_service/storage/repo"
)

type IStorage interface {
	Product() repo.ProductStorageI
}

type storagePg struct {
	db          *db.Postgres
	productRepo repo.ProductStorageI
}

func NewStoragePg(db *db.Postgres, log logger.Logger) *storagePg {
	return &storagePg{
		db:          db,
		productRepo: postgres.NewProductRepo(db, log),
	}
}

func (s storagePg) Product() repo.ProductStorageI {
	return s.productRepo
}
