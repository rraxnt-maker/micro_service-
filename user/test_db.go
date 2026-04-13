package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=localhost port=5430 user=useradmin password=secretpassword dbname=userdb sslmode=disable"
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open: %v", err)
	}
	defer db.Close()
	
	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping: %v", err)
	}
	
	log.Println("✅ Connected to PostgreSQL!")
	
	// Проверяем наличие таблицы
	var tableExists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_name = 'users'
	)`
	
	err = db.QueryRow(query).Scan(&tableExists)
	if err != nil {
		log.Fatalf("Failed to check table: %v", err)
	}
	
	if tableExists {
		log.Println("✅ Table 'users' exists")
	} else {
		log.Println("❌ Table 'users' does NOT exist")
	}
	
	// Пробуем вставить тестовую запись
	testID := "123e4567-e89b-12d3-a456-426614174000"
	_, err = db.Exec(`
		INSERT INTO users (id, email, username, full_name, age, created_at, updated_at)
		VALUES ($1, $2, '', '', 0, $3, $3)
		ON CONFLICT (id) DO NOTHING
	`, testID, "test@example.com", time.Now())
	
	if err != nil {
		log.Printf("Failed to insert test user: %v", err)
	} else {
		log.Println("✅ Test user inserted")
	}
}