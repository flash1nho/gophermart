package controllers

import (
    "fmt"
    "net/http"
    "encoding/json"
    "strconv"
    "time"

    h "github.com/flash1nho/go-musthave-diploma-tpl/app/helpers"

    "github.com/flash1nho/go-musthave-diploma-tpl/app/models"

    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
    "github.com/theplant/luhn"
)

type WithdrawnUserController struct {
    Pool *pgxpool.Pool
    Log  *zap.Logger
}

type WithdrawnUser struct {
    Current     *float64 `json:"current"`
    Withdrawn   *float64 `json:"withdrawn"`
    OrderNumber string   `json:"order"`
    Sum         *float64 `json:"sum"`
    ProcessedAt time.Time `json:"processed_at"`
}

func (controller *WithdrawnUserController) UserBalance(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    ctx := r.Context()
    userID := h.GetUserIDFromContext(ctx)
    withdrawnUser, err := models.UserBalance(ctx, userID, controller.Pool)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    responseWithdrawnUser := WithdrawnUser{
        Current: withdrawnUser.Current,
        Withdrawn: withdrawnUser.Withdrawn,
    }

    json.NewEncoder(w).Encode(responseWithdrawnUser)
    w.WriteHeader(http.StatusOK)
}

func (controller *WithdrawnUserController) UserBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var withdrawnUserRequest WithdrawnUser

    err := json.NewDecoder(r.Body).Decode(&withdrawnUserRequest)

    if err != nil {
        h.JSONError(w, "неверный формат запроса", http.StatusBadRequest)
        return
    }

    digitOrderNumber, err := strconv.Atoi(withdrawnUserRequest.OrderNumber)

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
    withdrawnUserBalance, err := models.UserBalance(ctx, userID, controller.Pool)

    if err != nil {
        h.JSONError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if *withdrawnUserBalance.Current < *withdrawnUserRequest.Sum {
        h.JSONError(w, "на счету недостаточно средств", http.StatusPaymentRequired)
        return
    }

    err = models.UserBalanceWithdraw(ctx, userID, withdrawnUserRequest.OrderNumber, withdrawnUserRequest.Sum, controller.Pool)

    if err != nil {
        h.JSONError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (controller *WithdrawnUserController) UserWithdrawals(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    ctx := r.Context()
    userID := h.GetUserIDFromContext(ctx)
    withdrawnUsers, err := models.UserWithdrawals(ctx, userID, controller.Pool)

    if len(withdrawnUsers) == 0 {
        h.JSONError(w, "нет ни одного списания", http.StatusNoContent)
        return
    }

    var responseWithdrawnUsers []WithdrawnUser

    for _, withdrawnUser := range withdrawnUsers {
        if err != nil {
            h.JSONError(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
            controller.Log.Error(fmt.Sprint(err))
            return
        }

        responseWithdrawnUser := WithdrawnUser{
            OrderNumber: withdrawnUser.OrderNumber,
            Sum: withdrawnUser.Withdrawn,
            ProcessedAt: withdrawnUser.ProcessedAt,
        }

        responseWithdrawnUsers = append(responseWithdrawnUsers, responseWithdrawnUser)
    }

    json.NewEncoder(w).Encode(responseWithdrawnUsers)
    w.WriteHeader(http.StatusOK)
}
