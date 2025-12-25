package models

import (
	  "context"
	  "errors"
	  "time"
	  "fmt"

		"github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/pgconn"
    "github.com/jackc/pgerrcode"
		"golang.org/x/crypto/bcrypt"
		"github.com/Masterminds/squirrel"
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

		sql, args, err := squirrel.Insert("users").
				Columns("login", "password", "created_at").
		    Values(login, hashedPassword, time.Now().UTC()).
		    Suffix("RETURNING id").
		    PlaceholderFormat(squirrel.Dollar).
		    ToSql()

		if err != nil {
			  fmt.Println(err)
				return 0, err
		}

		err = pool.QueryRow(ctx, sql, args...).Scan(&user.ID)

		if err != nil {
			  fmt.Println(err)
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
