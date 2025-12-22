package controllers

import (
    "fmt"
    "net/http"
    "encoding/json"
    "io"
    "time"
    "strconv"
    "context"

    "github.com/flash1nho/go-musthave-diploma-tpl/app/models"
    "github.com/flash1nho/go-musthave-diploma-tpl/app/helpers"
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
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    defer r.Body.Close()

    orderNumber := string(body)

    if orderNumber == "" {
        http.Error(w, "неверный формат запроса", http.StatusBadRequest)
        return
    }

    digitOrderNumber, err := strconv.Atoi(orderNumber)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if !luhn.Valid(digitOrderNumber) {
        http.Error(w, "неверный формат номера заказа", http.StatusUnprocessableEntity)
        return
    }

    ctx := r.Context()
    userID := helpers.GetUserIDFromContext(ctx)
    orderNumber, err = models.OrderCreate(ctx, orderNumber, userID, controller.Pool)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if orderNumber == "0" {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "номер заказа уже был загружен этим пользователем")
        return
    } else if orderNumber == "-1" {
        w.WriteHeader(http.StatusConflict)
        fmt.Fprintln(w, "номер заказа уже был загружен другим пользователем")
        return
    }

    orderChan := make(chan string, 10)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go workers.AccrualWorker(ctx, orderChan, controller.AccrualURL, userID, controller.Log, controller.Pool)

    orderChan <- orderNumber

    close(orderChan)
    time.Sleep(1 * time.Second)

    w.WriteHeader(http.StatusAccepted)
    fmt.Fprintln(w, "новый номер заказа принят в обработку")
}

func (controller *OrderController) GetOrders(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    ctx := r.Context()
    userID := helpers.GetUserIDFromContext(ctx)
    orders, err := models.OrderList(ctx, userID, controller.Pool)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if len(orders) == 0 {
        w.WriteHeader(http.StatusNoContent)
        fmt.Fprintln(w, "нет данных для ответа")
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
