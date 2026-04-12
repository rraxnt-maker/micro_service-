// model/user.go
package model

type User struct {
	ID        string `json:"id" db:"id"` // UUID как в auth сервисе
	Email     string `json:"email" db:"email"`
	Username  string `json:"username" db:"username"`
	Age       int    `json:"age" db:"age"`
	FullName  string `json:"full_name" db:"full_name"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
}
