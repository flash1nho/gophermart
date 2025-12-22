package models

import (
	  "context"
	  "errors"

		"github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgerrcode"
		"golang.org/x/crypto/bcrypt"
)

type User struct {
    ID       int
    Login    string
    Password string
}

func UserRegister(ctx context.Context, login string, password string, pool *pgxpool.Pool) (int, error) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		if err != nil {
				return 0, err
		}

		var user User

    query := `INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id`
		err = pool.QueryRow(ctx, query, login, hashedPassword).Scan(&user.ID)

		if err != nil {
        var pgErr *pgconn.PgError

        if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
						return 0, nil
        }

				return 0, err
		}

		return user.ID, nil
}

func UserLogin(ctx context.Context, login string, password string, pool *pgxpool.Pool) (int, error) {
		var user User

    query := `SELECT id, password FROM users WHERE users.login = $1`
		err := pool.QueryRow(ctx, query, login).Scan(&user.ID, &user.Password)

		if err != nil {
				return 0, err
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	  if err != nil {
				return 0, nil
		}

		return user.ID, nil
}
