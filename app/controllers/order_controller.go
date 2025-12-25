package controllers

import (
    "fmt"
    "net/http"
    "encoding/json"
    "io"
    "time"
    "strconv"

    h "github.com/flash1nho/go-musthave-diploma-tpl/app/helpers"

    "github.com/flash1nho/go-musthave-diploma-tpl/app/models"
    "github.com/flash1nho/go-musthave-diploma-tpl/app/workers"

    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
    "github.com/theplant/luhn"
)

type OrderController struct {
    Pool       *pgxpool.Pool
    Log        *zap.Logger
    AccrualURL string
}

type Order struct {
    Number     string             `json:"number"`
    Status     models.OrderStatus `json:"status"`
    Accrual    *float64           `json:"accrual,omitempty"`
    UploadedAt time.Time          `json:"uploaded_at"`
}

func (controller *OrderController) PostOrders(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)

    if err != nil {
        h.JSONError(w, "неверный формат запроса", http.StatusBadRequest)
        return
    }

    defer r.Body.Close()

    orderNumber := string(body)

    if orderNumber == "" {
        h.JSONError(w, "неверный формат запроса", http.StatusBadRequest)
        return
    }

    digitOrderNumber, err := strconv.Atoi(orderNumber)

    if err != nil {
        h.JSONError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if !luhn.Valid(digitOrderNumber) {
        h.JSONError(w, "неверный формат номера заказа", http.StatusUnprocessableEntity)
        return
    }

    ctx := r.Context()
    userID := h.GetUserIDFromContext(ctx)
    orderNumber, err = models.OrderCreate(ctx, orderNumber, userID, controller.Pool)

    if err != nil {
        h.JSONError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if orderNumber == "0" {
        h.JSONError(w, "номер заказа уже был загружен этим пользователем", http.StatusOK)
        return
    } else if orderNumber == "-1" {
        h.JSONError(w, "номер заказа уже был загружен другим пользователем", http.StatusConflict)
        return
    }

    workers.Run(controller.AccrualURL, userID, orderNumber, controller.Log, controller.Pool)

    h.JSONSuccess(w, "новый номер заказа принят в обработку", http.StatusAccepted)
}

func (controller *OrderController) GetOrders(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    ctx := r.Context()
    userID := h.GetUserIDFromContext(ctx)
    orders, err := models.OrderList(ctx, userID, controller.Pool)

    if err != nil {
        h.JSONError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if len(orders) == 0 {
        h.JSONError(w, "нет данных для ответа", http.StatusNoContent)
        return
    }

    var responseOrders []Order

    for _, order := range orders {
        responseOrder := Order{
            Number: order.Number,
            Status: order.Status,
            Accrual: order.Accrual,
            UploadedAt: order.UploadedAt,
        }

        responseOrders = append(responseOrders, responseOrder)
    }

    json.NewEncoder(w).Encode(responseOrders)
    w.WriteHeader(http.StatusOK)
}
