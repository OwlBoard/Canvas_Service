package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() (*pgxpool.Pool, error) {
	dbURL := "postgres://admin:admin@postgres_db:5432/canvas_db?sslmode=disable"
	if dbURL == "" {
		return nil, fmt.Errorf("la variable de entorno DATABASE_URL no está configurada")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de conexiones: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("no se pudo hacer ping a la base de datos: %w", err)
	}

	fmt.Println("¡Conexión exitosa con la base de datos PostgreSQL!")
	return pool, nil
}
