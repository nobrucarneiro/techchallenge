package main

import (
	configs "g37-lanchonete/configs"
	"g37-lanchonete/internal/api"
	"g37-lanchonete/internal/controllers/_api"
	"g37-lanchonete/internal/core/usecases"
	authorizerDriver "g37-lanchonete/internal/infra/drivers/auth"
	httpDriver "g37-lanchonete/internal/infra/drivers/http"
	paymentDriver "g37-lanchonete/internal/infra/drivers/payment"
	sqlDriver "g37-lanchonete/internal/infra/drivers/sql"
	"g37-lanchonete/internal/infra/gateways"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	config := configs.NewConfig()
	appConfig, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	httpClient := httpDriver.NewHttpClient()
	postgresSQLClient := createPostgresSQLClient(appConfig)
	err = performMigrations(postgresSQLClient)
	if err != nil {
		panic(err)
	}

	authorizerClient := httpDriver.NewHttpClient()
	authorizer := authorizerDriver.NewAuthorizer(authorizerClient, appConfig.AuthorizerURL)

	paymentBroker := paymentDriver.NewMercadoPagoBroker(httpClient, appConfig.PaymentBrokerURL)

	customerRepositoryGateway := gateways.NewCustomerRepositoryGateway(postgresSQLClient)
	productRepositoryGateway := gateways.NewProductRepositoryGateway(postgresSQLClient)
	orderRepositoryGateway := gateways.NewOrderRepositoryGateway(postgresSQLClient)

	customerUsecase := usecases.NewCustomerUsecase(customerRepositoryGateway)
	productUsecase := usecases.NewProductUsecase(productRepositoryGateway)
	paymentUsecase := usecases.NewPaymentUsecase(appConfig.NotificationURL, appConfig.SponsorId, paymentBroker)
	authorizerUsecase := usecases.NewAuthorizerUsecase(authorizer)
	orderUsecase := usecases.NewOrderUsecase(authorizerUsecase, paymentUsecase, productUsecase, orderRepositoryGateway)

	customerController := _api.NewCustomerController(customerUsecase)
	productController := _api.NewProductController(productUsecase)
	orderController := _api.NewOrderController(orderUsecase)

	apiParams := api.ApiParams{
		CustomerController: customerController,
		ProductController:  productController,
		OrderController:    orderController,
	}
	api := api.NewApi(apiParams)
	api.Run(":8080")
}

func createPostgresSQLClient(appConfig configs.AppConfig) sqlDriver.SQLClient {
	db, err := sqlDriver.NewPostgresSQLClient(appConfig.DatabaseUser, appConfig.DatabasePassword, appConfig.DatabaseHost, appConfig.DatabasePort, appConfig.DatabaseName)
	if err != nil {
		panic("failed to connect database")
	}

	err = db.Ping()
	if err != nil {
		panic("failed to ping database")
	}

	return db
}

func performMigrations(client sqlDriver.SQLClient) error {
	driver, err := postgres.WithInstance(client.GetConnection(), &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
