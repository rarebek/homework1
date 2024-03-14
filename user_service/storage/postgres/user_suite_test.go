package postgres

import (
	"EXAM3/user_service/config"
	pb "EXAM3/user_service/genproto/user_service"
	"EXAM3/user_service/pkg/db"
	"EXAM3/user_service/pkg/logger"
	"EXAM3/user_service/storage/repo"
	"context"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type UserRepositorySuiteTest struct {
	suite.Suite
	CleanupFunc func()
	Repository  repo.UserStorageI
}

func (u *UserRepositorySuiteTest) SetupSuite() {
	db, _ := db.New(*config.Load())
	u.Repository = NewUserRepo(db, logger.New("", ""))
	u.CleanupFunc = db.Close
}

func (p *UserRepositorySuiteTest) TestPositionCRUD() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(7))
	defer cancel()
	id := uuid.New().String()
	profile := randomdata.GenerateProfile(randomdata.RandomGender)

	//Create user
	user := &pb.User{
		Id:       id,
		Name:     randomdata.FullName(randomdata.RandomGender),
		Age:      int64(randomdata.Number(100)),
		Username: profile.Login.Username,
		Email:    randomdata.Email(),
		Password: profile.Login.Password,
	}
	createResp, err := p.Repository.Create(ctx, user)
	p.Suite.NoError(err)
	p.Suite.NotNil(createResp)

	//Get user
	userId := &pb.UserId{UserId: id}
	getResp, err := p.Repository.GetUserById(ctx, userId)
	p.Suite.NoError(err)
	p.Suite.NotNil(getResp)
	p.Suite.Equal(id, getResp.Id)
	p.Suite.Equal(user.Name, getResp.Name)
	p.Suite.Equal(user.Age, getResp.Age)
	p.Suite.Equal(user.Username, getResp.Username)
	p.Suite.Equal(user.Email, getResp.Email)
	p.Suite.Equal(user.Password, getResp.Password)

	//GetAll
	listResp, err := p.Repository.ListUsers(ctx, &pb.GetAllUserRequest{Page: 1, Limit: 10})
	p.Suite.NoError(err)
	p.Suite.NotNil(listResp)

	//Update user
	updatedName := randomdata.FullName(randomdata.RandomGender)
	user.Name = updatedName
	updateResp, err := p.Repository.UpdateUserById(ctx, user)
	p.Suite.NoError(err)
	p.Suite.NotNil(updateResp)
	p.Suite.Equal(updatedName, user.Name)

	//CheckField
	checkResp, err := p.Repository.CheckField(ctx, &pb.CheckFieldRequest{
		Field: "email",
		Data:  user.Email,
	})
	p.Suite.NoError(err)
	p.Suite.NotNil(checkResp)
	p.Suite.Equal(checkResp.Status, true)

	//Delete user
	_, err = p.Repository.Delete(ctx, userId)
	p.Suite.NoError(err)

}

func (p *UserRepositorySuiteTest) TearDownSuite() {
	p.CleanupFunc()
}

func TestCategoryRepository(t *testing.T) {
	suite.Run(t, new(UserRepositorySuiteTest))
}
