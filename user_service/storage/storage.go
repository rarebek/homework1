package storage

import (
	"EXAM3/user_service/pkg/db"
	"EXAM3/user_service/pkg/logger"
	"EXAM3/user_service/storage/postgres"
	"EXAM3/user_service/storage/repo"
)

type IStorage interface {
	User() repo.UserStorageI
}

type storagePg struct {
	db       *db.Postgres
	userRepo repo.UserStorageI
}

func NewStoragePg(db *db.Postgres, log logger.Logger) *storagePg {
	return &storagePg{
		db:       db,
		userRepo: postgres.NewUserRepo(db, log),
	}
}

func (s storagePg) User() repo.UserStorageI {
	return s.userRepo
}
