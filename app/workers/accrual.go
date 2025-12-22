package workers

import (
		"context"
		"encoding/json"
		"fmt"
		"net/http"
		"time"

		"github.com/flash1nho/go-musthave-diploma-tpl/app/models"

    "github.com/jackc/pgx/v5/pgxpool"
		"go.uber.org/zap"
)

type Order struct {
		Number  string  `json:"order"`
		Status  string  `json:"status"`
		Accrual *float64 `json:"accrual,omitempty"`
}

func AccrualWorker(ctx context.Context, orderChan <-chan string, accrualURL string, userID int, log *zap.Logger, pool *pgxpool.Pool) {
		client := &http.Client{
				Timeout: 5 * time.Second,
		}

		select {
		case <-ctx.Done():
				return
		case orderNumber, ok := <-orderChan:
				if !ok {
						return
				}

				processOrder(ctx, client, accrualURL, orderNumber, userID, log, pool)
		}
}

func processOrder(ctx context.Context, client *http.Client, baseURL string, orderNumber string, userID int, log *zap.Logger, pool *pgxpool.Pool) {
		url := fmt.Sprintf("%s/api/orders/%s", baseURL, orderNumber)
		resp, err := client.Get(url)

		if err != nil {
				log.Error(fmt.Sprintf("ошибка запроса заказа %s: %v", orderNumber, err))
				return
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
				var order Order

				if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
						log.Error(fmt.Sprint(err))
						return
				}

				err = models.UpdateUserOrder(ctx, userID, order.Number, order.Accrual, order.Status, pool)

				if err != nil {
						log.Error(fmt.Sprint(err))
						return
				}

				log.Info(fmt.Sprintf("заказ %s обработан.", order.Number))
		} else if resp.StatusCode == http.StatusTooManyRequests {
				log.Info("превышено количество запросов к сервису")
		} else if resp.StatusCode == http.StatusNoContent {
				log.Info("заказ не зарегистрирован в системе расчёта")
		}
}
