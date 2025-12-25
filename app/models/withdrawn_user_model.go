package models

import (
    "context"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type WithdrawnUser struct {
    OrderNumber string
    UserID      int
    Current     *float64
    Withdrawn   *float64
    ProcessedAt  time.Time
}

func UserBalance(ctx context.Context, userID int, pool *pgxpool.Pool) (WithdrawnUser, error) {
    var withdrawnUser WithdrawnUser

    query := `
      WITH withdrawn_users_cte AS (
        SELECT withdrawn_users.user_id,
               SUM(withdrawn_users.withdrawn) AS withdrawn
        FROM withdrawn_users
        WHERE withdrawn_users.user_id = $1
        GROUP BY withdrawn_users.user_id
      )
      SELECT users.balance,
             COALESCE(withdrawn_users_cte.withdrawn, 0)
      FROM users
      LEFT JOIN withdrawn_users_cte ON withdrawn_users_cte.user_id = users.id
      WHERE users.id = $1
    `

    err := pool.QueryRow(ctx, query, userID).Scan(&withdrawnUser.Current, &withdrawnUser.Withdrawn)

    if err != nil {
        return withdrawnUser, err
    }

	return withdrawnUser, nil
}

func UserBalanceWithdraw(ctx context.Context, userID int, orderNumber string, sum *float64, pool *pgxpool.Pool) error {
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

    query = `INSERT INTO withdrawn_users (withdrawn, user_id, order_number, processed_at) VALUES ($1, $2, $3, $4)`
    _, err = tx.Exec(ctx, query, sum, id, orderNumber, time.Now().UTC())

    if err != nil {
        return err
    }

    query = `UPDATE users SET balance = balance - $1 WHERE id = $2`
    _, err = tx.Exec(ctx, query, sum, id)

    if err != nil {
        return err
    }

    err = tx.Commit(ctx)

    if err != nil {
        return err
    }

    return nil
}

func UserWithdrawals(ctx context.Context, userID int, pool *pgxpool.Pool) ([]WithdrawnUser, error) {
    var withdrawnUsers []WithdrawnUser

    query := `SELECT order_number, withdrawn, processed_at FROM withdrawn_users WHERE user_id = $1 ORDER BY processed_at DESC`
    rows, err := pool.Query(ctx, query, userID)

    if err != nil {
        return nil, err
    }

    defer rows.Close()

    for rows.Next() {
        var withdrawnUser WithdrawnUser

        err = rows.Scan(&withdrawnUser.OrderNumber, &withdrawnUser.Withdrawn, &withdrawnUser.ProcessedAt)

        if err != nil {
            return nil, err
        }

        withdrawnUsers = append(withdrawnUsers, withdrawnUser)
    }

    return withdrawnUsers, nil
}
