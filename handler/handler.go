package handler

import (
    "github.com/flash1nho/go-musthave-diploma-tpl/app/controllers"

    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

type Handler struct {
    UserController          *controllers.UserController
    OrderController         *controllers.OrderController
    WithdrawnUserController *controllers.WithdrawnUserController
    Log                     *zap.Logger
}

func NewHandler(pool *pgxpool.Pool, log *zap.Logger, accrualURL string) *Handler {
    return &Handler{
        UserController: &controllers.UserController{Pool: pool, Log: log},
        OrderController: &controllers.OrderController{Pool: pool, Log: log, AccrualURL: accrualURL},
        WithdrawnUserController: &controllers.WithdrawnUserController{Pool: pool, Log: log},
        Log: log,
    }
}
