package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"user/internal/config"
	"user/internal/model"

	_ "github.com/lib/pq"
)

type Storage interface {
	GetUser(id string) (*model.User, error)
	GetPublicUser(id string) (*model.PublicUser, error)
	CreateOrUpdateUser(id, email string) (string, error)
	UpdateUser(id string, updates map[string]interface{}) (*model.User, error)
	DeleteUser(id string) (bool, error)
	CountUsers() (int, error)
	Ping() error
	Close() error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	log.Printf("Connecting to PostgreSQL: host=%s port=%s user=%s dbname=%s", 
		config.DBHost, config.DBPort, config.DBUser, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Проверяем соединение
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Connected to PostgreSQL successfully!")
	
	// Проверяем наличие таблицы
	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'users'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		log.Printf("Warning: could not check if table exists: %v", err)
	} else if !tableExists {
		log.Println("⚠️  Table 'users' does not exist. Creating...")
		if err := createTables(db); err != nil {
			return nil, fmt.Errorf("failed to create tables: %v", err)
		}
		log.Println("✅ Tables created successfully!")
	} else {
		log.Println("✅ Table 'users' exists")
	}

	return &PostgresStorage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) NOT NULL,
			username VARCHAR(50),
			full_name VARCHAR(100),
			age INTEGER CHECK (age >= 0 AND age <= 150),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		
		CREATE TRIGGER update_users_updated_at 
			BEFORE UPDATE ON users 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();
	`
	
	_, err := db.Exec(query)
	return err
}

func (s *PostgresStorage) GetUser(id string) (*model.User, error) {
	var user model.User
	query := `SELECT id, email, username, full_name, age, created_at, updated_at 
	          FROM users WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.Age,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

func (s *PostgresStorage) GetPublicUser(id string) (*model.PublicUser, error) {
	var user model.PublicUser
	query := `SELECT id, username, full_name, age, created_at 
	          FROM users WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.FullName,
		&user.Age,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get public user: %v", err)
	}

	return &user, nil
}

// Исправленная функция CreateOrUpdateUser
func (s *PostgresStorage) CreateOrUpdateUser(id, email string) (string, error) {
	log.Printf("Creating/updating user: id=%s email=%s", id, email)
	
	// Сначала проверяем, существует ли пользователь
	var existingEmail string
	var userExists bool
	
	err := s.db.QueryRow("SELECT email FROM users WHERE id = $1", id).Scan(&existingEmail)
	if err == nil {
		userExists = true
	} else if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to check user: %v", err)
	}
	
	if userExists {
		// Пользователь существует - обновляем email только если он изменился
		if existingEmail != email {
			// Проверяем, не занят ли новый email другим пользователем
			var count int
			err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2", email, id).Scan(&count)
			if err != nil {
				return "", fmt.Errorf("failed to check email: %v", err)
			}
			if count > 0 {
				return "", fmt.Errorf("email already taken")
			}
			
			_, err = s.db.Exec("UPDATE users SET email = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", email, id)
			if err != nil {
				return "", fmt.Errorf("failed to update user: %v", err)
			}
			log.Printf("User updated: %s", id)
		}
		return "updated", nil
	}
	
	// Проверяем, не занят ли email
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("failed to check email: %v", err)
	}
	if count > 0 {
		return "", fmt.Errorf("email already taken")
	}
	
	// Создаем нового пользователя
	now := time.Now()
	_, err = s.db.Exec(`
		INSERT INTO users (id, email, username, full_name, age, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, email, "", "", 0, now, now)
	
	if err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}
	
	log.Printf("User created: %s", id)
	return "created", nil
}

func (s *PostgresStorage) UpdateUser(id string, updates map[string]interface{}) (*model.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %v", err)
	}
	if !exists {
		return nil, nil
	}

	query := "UPDATE users SET "
	var args []interface{}
	argCounter := 1

	if username, ok := updates["username"].(string); ok {
		query += fmt.Sprintf("username = $%d, ", argCounter)
		args = append(args, username)
		argCounter++
	}

	if fullName, ok := updates["full_name"].(string); ok {
		query += fmt.Sprintf("full_name = $%d, ", argCounter)
		args = append(args, fullName)
		argCounter++
	}

	if age, ok := updates["age"].(int); ok {
		query += fmt.Sprintf("age = $%d, ", argCounter)
		args = append(args, age)
		argCounter++
	}

	if argCounter == 1 {
		return nil, fmt.Errorf("no fields to update")
	}

	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d", argCounter)
	args = append(args, id)

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	var user model.User
	selectQuery := `SELECT id, email, username, full_name, age, created_at, updated_at 
	                FROM users WHERE id = $1`

	err = tx.QueryRow(selectQuery, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FullName,
		&user.Age,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &user, nil
}

func (s *PostgresStorage) DeleteUser(id string) (bool, error) {
	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return false, fmt.Errorf("failed to delete user: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected > 0, nil
}

func (s *PostgresStorage) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %v", err)
	}
	return count, nil
}

func (s *PostgresStorage) Ping() error {
	return s.db.Ping()
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

var DB Storage

func Init() error {
	var err error
	DB, err = NewPostgresStorage()
	return err
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}