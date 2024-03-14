package postgres

import (
	pb "EXAM3/product_service/genproto/product_service"
	"EXAM3/product_service/pkg/db"
	"EXAM3/product_service/pkg/logger"
	"context"
	"time"

	"github.com/Masterminds/squirrel"
)

type productRepo struct {
	db  *db.Postgres
	log logger.Logger
}

func NewProductRepo(db *db.Postgres, log logger.Logger) *productRepo {
	return &productRepo{
		db:  db,
		log: log,
	}
}

func (r *productRepo) CreateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	query := r.db.Builder.Insert("products").
		Columns(`
	  id, name, description, price, amount
	  `).
		Values(
			req.Id, req.Name, req.Description, req.Price, req.Amount,
		).
		Suffix("RETURNING created_at")

	err := query.RunWith(r.db.DB).QueryRow().Scan(&req.CreatedAt)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (r *productRepo) GetProductById(ctx context.Context, req *pb.ProductId) (*pb.Product, error) {
	respProduct := &pb.Product{}

	query := r.db.Builder.Select(`
	  id, name, description, price, amount, created_at
	`).From("products").Where(squirrel.Eq{"id": req.ProductId})

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&respProduct.Id,
		&respProduct.Name,
		&respProduct.Description,
		&respProduct.Price,
		&respProduct.Amount,
		&respProduct.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return respProduct, nil
}

func (r *productRepo) ListProducts(ctx context.Context, req *pb.GetAllProductRequest) (*pb.GetAllProductResponse, error) {
	var (
		respProducts = &pb.GetAllProductResponse{Count: 0}
	)

	query := r.db.Builder.Select(
		`id, name, description, price, amount, created_at
	`).From("products")

	query = query.Offset(uint64((req.Page - 1) * req.Limit)).Limit(uint64(req.Limit))

	rows, err := query.RunWith(r.db.DB).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		respProduct := &pb.Product{}
		err = rows.Scan(
			&respProduct.Id,
			&respProduct.Name,
			&respProduct.Description,
			&respProduct.Price,
			&respProduct.Amount,
			&respProduct.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		respProducts.Products = append(respProducts.Products, respProduct)
		respProducts.Count++
	}

	return respProducts, nil
}

func (r *productRepo) IncreaseProductAmount(ctx context.Context, req *pb.ProductAmountRequest) (*pb.ProductAmountResponse, error) {
	var (
		response  = &pb.ProductAmountResponse{Product: &pb.Product{}}
		updateMap = make(map[string]interface{})
		where     = squirrel.And{squirrel.Eq{"id": req.ProductId}}
	)

	var currentAmount int64
	query := r.db.Builder.Select("amount").From("products").Where(squirrel.Eq{"id": req.ProductId})

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&currentAmount,
	)
	if err != nil {
		return &pb.ProductAmountResponse{
			IsEnough: false,
			Product:  nil,
		}, err
	}

	updateMap["amount"] = req.Amount + currentAmount
	updateMap["updated_at"] = time.Now()
	response.Product.Amount = req.Amount + currentAmount

	query2 := r.db.Builder.Update("products").SetMap(updateMap).
		Where(where).
		Suffix("RETURNING id, name, description, price, amount, created_at")

	err = query2.RunWith(r.db.DB).QueryRow().Scan(
		&response.Product.Id,
		&response.Product.Name,
		&response.Product.Description,
		&response.Product.Price,
		&response.Product.Amount,
		&response.Product.CreatedAt,
	)
	if err != nil {
		return &pb.ProductAmountResponse{
			IsEnough: false,
			Product:  nil,
		}, err
	}

	response.IsEnough = true

	return response, nil
}

func (r *productRepo) DecreaseProductAmount(ctx context.Context, req *pb.ProductAmountRequest) (*pb.ProductAmountResponse, error) {
	var (
		response  = &pb.ProductAmountResponse{Product: &pb.Product{}}
		updateMap = make(map[string]interface{})
		where     = squirrel.And{squirrel.Eq{"id": req.ProductId}}
	)

	response.IsEnough = true
	var currentAmount int64
	query := r.db.Builder.Select("amount").From("products").Where(squirrel.Eq{"id": req.ProductId})

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&currentAmount,
	)
	if err != nil {
		response.IsEnough = false
		return response, err
	}

	if currentAmount == 0 {
		response.IsEnough = false
		return response, nil
	}

	changeTo := currentAmount - req.Amount
	if changeTo < 0 {
		changeTo = req.Amount
		response.IsEnough = false
	}

	updateMap["amount"] = changeTo
	updateMap["updated_at"] = time.Now()

	query2 := r.db.Builder.Update("products").SetMap(updateMap).
		Where(where).
		Suffix("RETURNING id, name, description, price, amount, created_at")

	err = query2.RunWith(r.db.DB).QueryRow().Scan(
		&response.Product.Id,
		&response.Product.Name,
		&response.Product.Description,
		&response.Product.Price,
		&response.Product.Amount,
		&response.Product.CreatedAt,
	)
	if err != nil {
		response.IsEnough = false
		return response, err
	}
	response.Product.Amount = changeTo

	return response, nil
}

func (r *productRepo) UpdateProduct(ctx context.Context, req *pb.Product) (*pb.Product, error) {
	var (
		mp             = make(map[string]interface{})
		whereCondition = squirrel.And{squirrel.Eq{"id": req.Id}}
	)

	mp["name"] = req.Name
	mp["description"] = req.Description
	mp["price"] = req.Price
	mp["amount"] = req.Amount
	mp["updated_at"] = time.Now()
	query := r.db.Builder.Update("products").SetMap(mp).Where(
		whereCondition,
	).Suffix(`
		RETURNING id, name, description, price, amount
	`)

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&req.Id,
		&req.Name,
		&req.Description,
		&req.Price,
		&req.Amount,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *productRepo) DeleteProduct(ctx context.Context, req *pb.ProductId) (*pb.Status, error) {
	query := r.db.Builder.Delete("products").Where(
		squirrel.Eq{"id": req.ProductId},
	)

	_, err := query.RunWith(r.db.DB).Exec()
	if err != nil {
		return &pb.Status{
			Success: false,
		}, err
	}

	return &pb.Status{
		Success: true,
	}, nil
}

func (r *productRepo) CheckAmount(ctx context.Context, req *pb.ProductId) (*pb.CheckAmountResponse, error) {
	var checkResult pb.CheckAmountResponse
	query := r.db.Builder.Select("amount").From("products").Where(
		squirrel.Eq{"id": req.ProductId},
	)

	err := query.RunWith(r.db.DB).QueryRow().Scan(
		&checkResult.Amount,
	)
	checkResult.ProductId = req.ProductId
	if err != nil {
		return nil, err
	}

	return &checkResult, nil
}

func (r *productRepo) BuyProduct(ctx context.Context, req *pb.BuyProductRequest) (*pb.Product, error) {
	query := r.db.Builder.Insert("users_products").
		Columns("user_id, product_id, amount").Values(
		req.UserId,
		req.ProductId,
		req.Amount,
	)

	_, err := query.RunWith(r.db.DB).Exec()
	if err != nil {
		return nil, err
	}

	product, err := r.GetProductById(ctx, &pb.ProductId{
		ProductId: req.ProductId,
	})

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepo) GetBoughtProductsByUserId(ctx context.Context, req *pb.UserId) (*pb.GetBoughtProductsResponse, error) {
	query := r.db.Builder.Select("product_id").
		From("users_products").Where(squirrel.Eq{"user_id": req.UserId})
	rows, err := query.RunWith(r.db.DB).Query()
	if err != nil {
		return nil, err
	}
	var products []*pb.Product

	for rows.Next() {
		var productId string

		if err := rows.Scan(&productId); err != nil {
			return nil, err
		}

		respProduct, err := r.GetProductById(ctx, &pb.ProductId{ProductId: productId})
		if err != nil {
			return nil, err
		}

		products = append(products, respProduct)
	}

	return &pb.GetBoughtProductsResponse{
		Products: products,
	}, nil
}
