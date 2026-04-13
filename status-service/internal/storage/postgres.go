package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"status-service/internal/config"
	"status-service/internal/model"

	_ "github.com/lib/pq"
)

type Storage interface {
	SetStatus(userID string, req *model.SetStatusRequest) (*model.Status, error)
	GetStatus(userID string) (*model.Status, error)
	GetPublicStatus(userID string) (*model.PublicStatus, error)
	GetBatchStatuses(userIDs []string) ([]*model.PublicStatus, error)
	DeleteStatus(userID string) error
	SetDND(userID string, duration int) (*model.Status, error)
	GetHistory(userID string, limit, offset int) ([]*model.StatusHistory, int, error)
	SyncUser(userID string) error
	DeleteUser(userID string) error
	DeleteExpiredStatuses() (int64, error)
	Ping() error
	Close() error
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("✅ Connected to PostgreSQL!")
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) SetStatus(userID string, req *model.SetStatusRequest) (*model.Status, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Сохраняем в историю
	_, err = tx.Exec(`
		INSERT INTO status_history (user_id, text, emoji)
		SELECT $1, text, emoji FROM statuses WHERE user_id = $1
	`, userID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &t
	}

	statusType := req.Type
	if statusType == "" {
		statusType = "normal"
	}

	_, err = tx.Exec(`
		INSERT INTO statuses (user_id, text, emoji, type, activity, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			text = EXCLUDED.text,
			emoji = EXCLUDED.emoji,
			type = EXCLUDED.type,
			activity = EXCLUDED.activity,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`, userID, req.Text, req.Emoji, statusType, req.Activity, expiresAt)

	if err != nil {
		return nil, err
	}

	var status model.Status
	err = tx.QueryRow(`
		SELECT user_id, text, emoji, type, activity, expires_at, created_at, updated_at
		FROM statuses WHERE user_id = $1
	`, userID).Scan(&status.UserID, &status.Text, &status.Emoji, &status.Type,
		&status.Activity, &status.ExpiresAt, &status.CreatedAt, &status.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &status, nil
}

func (s *PostgresStorage) GetStatus(userID string) (*model.Status, error) {
	var status model.Status
	err := s.db.QueryRow(`
		SELECT user_id, text, emoji, type, activity, expires_at, created_at, updated_at
		FROM statuses WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`, userID).Scan(&status.UserID, &status.Text, &status.Emoji, &status.Type,
		&status.Activity, &status.ExpiresAt, &status.CreatedAt, &status.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *PostgresStorage) GetPublicStatus(userID string) (*model.PublicStatus, error) {
	var status model.PublicStatus
	err := s.db.QueryRow(`
		SELECT user_id, text, emoji, type, activity, expires_at
		FROM statuses WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`, userID).Scan(&status.UserID, &status.Text, &status.Emoji, &status.Type,
		&status.Activity, &status.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *PostgresStorage) GetBatchStatuses(userIDs []string) ([]*model.PublicStatus, error) {
	if len(userIDs) == 0 {
		return []*model.PublicStatus{}, nil
	}

	query := `SELECT user_id, text, emoji, type, activity, expires_at 
	          FROM statuses WHERE user_id = ANY($1) AND (expires_at IS NULL OR expires_at > NOW())`
	
	rows, err := s.db.Query(query, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*model.PublicStatus
	for rows.Next() {
		var status model.PublicStatus
		err := rows.Scan(&status.UserID, &status.Text, &status.Emoji, &status.Type,
			&status.Activity, &status.ExpiresAt)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, &status)
	}

	return statuses, nil
}

func (s *PostgresStorage) DeleteStatus(userID string) error {
	_, err := s.db.Exec("DELETE FROM statuses WHERE user_id = $1", userID)
	return err
}

func (s *PostgresStorage) SetDND(userID string, duration int) (*model.Status, error) {
	req := &model.SetStatusRequest{
		Text:     "Не беспокоить",
		Emoji:    "🔕",
		Type:     "dnd",
		Activity: "",
	}
	if duration > 0 {
		req.ExpiresIn = duration
	}
	return s.SetStatus(userID, req)
}

func (s *PostgresStorage) GetHistory(userID string, limit, offset int) ([]*model.StatusHistory, int, error) {
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM status_history WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(`
		SELECT id, user_id, text, emoji, created_at
		FROM status_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var history []*model.StatusHistory
	for rows.Next() {
		var h model.StatusHistory
		err := rows.Scan(&h.ID, &h.UserID, &h.Text, &h.Emoji, &h.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		history = append(history, &h)
	}

	return history, total, nil
}

func (s *PostgresStorage) SyncUser(userID string) error {
	// Просто проверяем что пользователь может иметь статус, ничего не создаём
	return nil
}

func (s *PostgresStorage) DeleteUser(userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM statuses WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM status_history WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStorage) DeleteExpiredStatuses() (int64, error) {
	result, err := s.db.Exec("DELETE FROM statuses WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
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