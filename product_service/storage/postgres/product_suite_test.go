package postgres

import (
	"EXAM3/product_service/config"
	pb "EXAM3/product_service/genproto/product_service"
	"EXAM3/product_service/pkg/db"
	"EXAM3/product_service/pkg/logger"
	"EXAM3/product_service/storage/repo"
	"context"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ProductRepositorySuiteTest struct {
	suite.Suite
	CleanupFunc func()
	Repository  repo.ProductStorageI
}

func (u *ProductRepositorySuiteTest) SetupSuite() {
	db, _ := db.New(*config.Load())
	u.Repository = NewProductRepo(db, logger.New("", ""))
	u.CleanupFunc = db.Close
}

func (p *ProductRepositorySuiteTest) TestPositionCRUD() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(7))
	defer cancel()
	id := uuid.New().String()
	amount := int64(randomdata.Number(5, 999))

	//Create product
	product := &pb.Product{
		Id:          id,
		Name:        randomdata.FullName(randomdata.RandomGender),
		Description: randomdata.Title(randomdata.RandomGender),
		Price:       float32(randomdata.Number(999)),
		Amount:      amount,
	}
	createResp, err := p.Repository.CreateProduct(ctx, product)
	p.Suite.NoError(err)
	p.Suite.NotNil(createResp)

	//Get product
	productId := &pb.ProductId{ProductId: id}
	getResp, err := p.Repository.GetProductById(ctx, productId)
	p.Suite.NoError(err)
	p.Suite.NotNil(getResp)
	p.Suite.Equal(id, getResp.Id)
	p.Suite.Equal(product.Name, getResp.Name)
	p.Suite.Equal(product.Description, getResp.Description)
	p.Suite.Equal(product.Price, getResp.Price)
	p.Suite.Equal(product.Amount, getResp.Amount)

	//GetAll
	listResp, err := p.Repository.ListProducts(ctx, &pb.GetAllProductRequest{Page: 1, Limit: 10})
	p.Suite.NoError(err)
	p.Suite.NotNil(listResp)

	//Update product
	updatedName := randomdata.FullName(randomdata.RandomGender)
	product.Name = updatedName
	updateResp, err := p.Repository.UpdateProduct(ctx, product)
	p.Suite.NoError(err)
	p.Suite.NotNil(updateResp)
	p.Suite.Equal(updatedName, product.Name)

	//Check amount
	checkResp, err := p.Repository.CheckAmount(ctx, productId)
	p.Suite.NoError(err)
	p.Suite.NotNil(checkResp)
	p.Suite.Equal(checkResp.ProductId, productId.ProductId)
	p.Suite.Equal(checkResp.Amount, amount)

	//Buy product
	respProduct, err := p.Repository.BuyProduct(ctx, &pb.BuyProductRequest{
		UserId:    uuid.New().String(),
		ProductId: productId.ProductId,
		Amount:    1,
	})
	p.Suite.NoError(err)
	p.Suite.NotNil(respProduct)
	p.Suite.Equal(respProduct.Name, product.Name)
	p.Suite.Equal(respProduct.Description, product.Description)

	//Decrease
	resp, err := p.Repository.DecreaseProductAmount(ctx, &pb.ProductAmountRequest{
		ProductId: productId.ProductId,
		Amount:    1,
	})
	p.Suite.NoError(err)
	p.Suite.NotNil(resp)
	p.Suite.NotEqual(resp.Product.Amount, product.Amount)
	p.Suite.Equal(resp.IsEnough, true)
	p.Suite.Equal(resp.Product.Price, product.Price)

	//Increase product
	response, err := p.Repository.IncreaseProductAmount(ctx, &pb.ProductAmountRequest{
		ProductId: productId.ProductId,
		Amount:    1,
	})
	p.Suite.NoError(err)
	p.Suite.NotNil(response)
	p.Suite.Equal(response.IsEnough, true)
	p.Suite.Equal(response.Product.Price, product.Price)

	//Delete product
	_, err = p.Repository.DeleteProduct(ctx, productId)
	p.Suite.NoError(err)

}

func (p *ProductRepositorySuiteTest) TearDownSuite() {
	p.CleanupFunc()
}

func TestCategoryRepository(t *testing.T) {
	suite.Run(t, new(ProductRepositorySuiteTest))
}
