package db

import (
		"context"
		"fmt"

		"github.com/jackc/pgx/v5/pgxpool"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewDB(databaseURI string) (*pgxpool.Pool, error) {
	  var pool *pgxpool.Pool = nil
	  var err error

    if databaseURI == "" {
    		return nil, fmt.Errorf("не указаны реквизиты для подключения к БД: %w", err)
    } else {
			  pool, err = connect(databaseURI)

			  if err != nil {
			  		return nil, err
			  }

		    m, err := migrate.New("file://migrations", databaseURI)

		    if err != nil {
		    	  return nil, fmt.Errorf("ошибка загрузки миграций: %w", err)
		    }

		    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		        return nil, fmt.Errorf("ошибка запуска миграций: %w", err)
		    }
	  }

		return pool, nil
}

func connect(databaseURI string) (*pgxpool.Pool, error) {
		pool, err := pgxpool.New(context.Background(), databaseURI)

		if err != nil {
				return nil, fmt.Errorf("ошибка подключения к базе данных: %w", err)
		}

		err = pool.Ping(context.TODO())

		if err != nil {
				return nil, fmt.Errorf("ошибка пинга базы данных: %w", err)
		}

		return pool, nil
}
