package postgres

import (
	pb "EXAM3/user_service/genproto/user_service"
	"EXAM3/user_service/pkg/db"
	"EXAM3/user_service/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
)

type userRepo struct {
	db  *db.Postgres
	log logger.Logger
}

func NewUserRepo(db *db.Postgres, log logger.Logger) *userRepo {
	return &userRepo{
		db:  db,
		log: log,
	}
}

func (r *userRepo) Create(ctx context.Context, req *pb.User) (*pb.User, error) {
	var createdUser pb.User
	query := r.db.Builder.Insert("users").
		Columns(`
			id, name, age, username, email, password, refresh_token
		`).Values(
		req.Id, req.Name, req.Age, req.Username, req.Email, req.Password, req.RefreshToken,
	).Suffix("RETURNING id, name, age, username, email, password, refresh_token")

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&createdUser.Id,
		&createdUser.Name,
		&createdUser.Age,
		&createdUser.Username,
		&createdUser.Email,
		&createdUser.Password,
		&createdUser.RefreshToken,
	)
	if err != nil {
		return nil, err
	}
	return &createdUser, nil
}

func (r *userRepo) GetUserByUsername(ctx context.Context, req *pb.Username) (*pb.User, error) {
	var user pb.User

	query := r.db.Builder.Select(`
		id, name, age, username, email, password, refresh_token
	`).From(`
		users
	`)

	if req.Username != "" {
		query = query.Where(squirrel.Eq{"username": req.Username})
	} else {
		return nil, fmt.Errorf("username is required")
	}

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&user.Id,
		&user.Name,
		&user.Age,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) GetUserById(ctx context.Context, req *pb.UserId) (*pb.User, error) {
	var user pb.User

	query := r.db.Builder.Select(`
		id, name, age, username, email, password, refresh_token
	`).From(`
		users
	`)

	if req.UserId != "" {
		query = query.Where(squirrel.Eq{"id": req.UserId})
	} else {
		return nil, fmt.Errorf("user id is required")
	}

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&user.Id,
		&user.Name,
		&user.Age,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) UpdateUserById(ctx context.Context, req *pb.User) (*pb.User, error) {
	var user pb.User
	var (
		mp             = make(map[string]interface{})
		whereCondition = squirrel.And{squirrel.Eq{"id": req.Id}}
	)

	mp["name"] = req.Name
	mp["age"] = req.Age
	mp["updated_at"] = time.Now()

	query := r.db.Builder.Update("users").SetMap(mp).Where(
		whereCondition,
	).Suffix(`
		RETURNING id, name, age, username, email, password, refresh_token
	`)

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&user.Id,
		&user.Name,
		&user.Age,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepo) Delete(ctx context.Context, req *pb.UserId) (*pb.Empty, error) {
	query := r.db.Builder.Delete("users").Where(
		squirrel.Eq{"id": req.UserId},
	)
	_, err := query.RunWith(r.db.DB).Exec()
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (u *userRepo) ListUsers(ctx context.Context, req *pb.GetAllUserRequest) (*pb.GetAllUserResponse, error) {
	var (
		respUsers = &pb.GetAllUserResponse{Count: 0}
	)

	query := u.db.Builder.Select(`
		id, name, age, username, email, password, refresh_token
	`).From("users")

	query = query.Offset(uint64((req.Page - 1) * req.Limit)).Limit(uint64(req.Limit))

	rows, err := query.RunWith(u.db.DB).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		respUser := &pb.User{}
		err = rows.Scan(
			&respUser.Id,
			&respUser.Name,
			&respUser.Age,
			&respUser.Email,
			&respUser.Username,
			&respUser.Password,
			&respUser.RefreshToken,
		)
		if err != nil {
			return nil, err
		}
		respUsers.Users = append(respUsers.Users, respUser)
		respUsers.Count++
	}

	return respUsers, nil

}

func (u *userRepo) CheckField(ctx context.Context, req *pb.CheckFieldRequest) (*pb.CheckFieldResponse, error) {
	var (
		response = &pb.CheckFieldResponse{}
	)
	var resp int
	num := u.db.Builder.Select("count(1)").From("users").Where(squirrel.Eq{req.Field: req.Data})

	err := num.RunWith(u.db.DB).Scan(&resp)
	if err != nil {
		response.Status = false
		return response, err
	}
	if resp == 1 {
		response.Status = true
	} else if resp == 0 {
		response.Status = false
	}

	return response, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, req *pb.Email) (*pb.User, error) {
	var user pb.User
	query := r.db.Builder.Select(`
		id, name, age, username, email, password, refresh_token
	`).From(`
		users
	`)

	if req.Email != "" {
		query = query.Where(squirrel.Eq{"email": req.Email})
	} else {
		return nil, fmt.Errorf("user id is required")
	}

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&user.Id,
		&user.Name,
		&user.Age,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RefreshToken,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
