package models

import (
    "context"
    "errors"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgerrcode"
)

type OrderStatus string

const (
    StatusNew        OrderStatus = "NEW" // заказ загружен в систему, но не попал в обработку
    StatusRegistered OrderStatus = "REGISTERED" // заказ зарегистрирован, но вознаграждение не рассчитано
    StatusProcessing OrderStatus = "PROCESSING" // вознаграждение за заказ рассчитывается
    StatusInvalid    OrderStatus = "INVALID"    // система расчёта вознаграждений отказала в расчёте
    StatusProcessed  OrderStatus = "PROCESSED"  // данные по заказу проверены и информация о расчёте успешно получена
)

type Order struct {
    Number  	 string
    UserID  	 int
    Status  	 OrderStatus
    Accrual 	 *float64
    UploadedAt time.Time
}

func OrderCreate(ctx context.Context, number string, userID int, pool *pgxpool.Pool) (string, error) {
    var orderExists bool

    query := `SELECT EXISTS(SELECT 1 FROM orders WHERE number = $1 AND user_id != $2)`
    err := pool.QueryRow(ctx, query, number, userID).Scan(&orderExists)

    if err != nil {
        return "0", err
    } else if orderExists {
        return "-1", nil
    }

    var order Order

    query = `INSERT INTO orders (number, user_id, uploaded_at) VALUES ($1, $2, $3) RETURNING number`
    err = pool.QueryRow(ctx, query, number, userID, time.Now().UTC()).Scan(&order.Number)

    if err != nil {
        var pgErr *pgconn.PgError

        if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
            return "0", nil
        }

        return "0", err
    }

    return order.Number,  nil
}

func OrderList(ctx context.Context, userID int, pool *pgxpool.Pool) ([]Order, error) {
    var orders []Order

    query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1`
    rows, err := pool.Query(ctx, query, userID)

    if err != nil {
        return nil, err
    }

    defer rows.Close()

    for rows.Next() {
        var order Order

        err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)

        if err != nil {
            return nil, err
        }

        orders = append(orders, order)
    }

    return orders, nil
}

func UpdateUserOrder(ctx context.Context, userID int, orderNumber string, accrual *float64, status string, pool *pgxpool.Pool) error {
    tx, err := pool.Begin(ctx)

    if err != nil {
        return err
    }

    defer tx.Rollback(ctx)

    var id int

    query := `SELECT id FROM users WHERE id = $1 FOR UPDATE SKIP LOCKED`
    err = tx.QueryRow(ctx, query, userID).Scan(&id)

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil
        }

        return err
    }

    query = `UPDATE orders SET accrual = $1, status = $2 WHERE number = $3`
    _, err = tx.Exec(ctx, query, accrual, status, orderNumber)

    if err != nil {
        return err
    }

    if accrual != nil {
        query = `UPDATE users SET balance = balance + $1 WHERE id = $2`
        _, err = tx.Exec(ctx, query, accrual, id)

        if err != nil {
            return err
        }
    }

    err = tx.Commit(ctx)

    if err != nil {
        return err
    }

    return nil
}
