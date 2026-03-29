package manager

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Product struct {
	ID      int
	Model   string
	Company string
	Price   int
}

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string // Не храним в БД, используем для ввода
	CreatedAt time.Time
}

type Session struct {
	ID           int
	UserID       int
	SessionToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

var DB *sql.DB

// Инициализация БД
func InitDB() error {
	connStr := "postgres://postgres:1@localhost:5432/test_db?sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %v", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	fmt.Println("Успешное подключение к PostgreSQL!")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// Генерация случайного токена
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Создание нового пользователя
func CreateUser(username, email, password string) error {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля: %v", err)
	}

	_, err = DB.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3)",
		username, email, string(hashedPassword),
	)
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %v", err)
	}
	return nil
}

// Проверка учетных данных пользователя
func AuthenticateUser(username, password string) (*User, error) {
	var user User
	var passwordHash string

	err := DB.QueryRow(
		"SELECT id, username, email, password_hash FROM users WHERE username=$1",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка аутентификации: %v", err)
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("неверный пароль")
	}

	return &user, nil
}

// Создание сессии
func CreateSession(userID int) (string, error) {
	// Генерируем токен
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	// Устанавливаем время истечения (например, через 24 часа)
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = DB.Exec(
		"INSERT INTO sessions (user_id, session_token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("ошибка создания сессии: %v", err)
	}

	return token, nil
}

// Получение пользователя по токену сессии
func GetUserBySessionToken(token string) (*User, error) {
	var user User
	var expiresAt time.Time

	err := DB.QueryRow(`
        SELECT u.id, u.username, u.email, s.expires_at 
        FROM users u 
        JOIN sessions s ON u.id = s.user_id 
        WHERE s.session_token = $1 AND s.expires_at > NOW()`,
		token,
	).Scan(&user.ID, &user.Username, &user.Email, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("сессия не найдена или истекла")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
	}

	return &user, nil
}

// Удаление сессии (выход)
func DeleteSession(token string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE session_token=$1", token)
	if err != nil {
		return fmt.Errorf("ошибка удаления сессии: %v", err)
	}
	return nil
}

// Удаление всех сессий пользователя (опционально)
func DeleteAllUserSessions(userID int) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE user_id=$1", userID)
	return err
}

// Очистка просроченных сессий (можно запускать периодически)
func CleanExpiredSessions() error {
	_, err := DB.Exec("DELETE FROM sessions WHERE expires_at <= NOW()")
	return err
}

// Продукты (существующие функции)
func GetAllProducts() ([]Product, error) {
	var posts []Product
	rows, err := DB.Query("SELECT id, model, company, price FROM products")
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Model, &product.Company, &product.Price)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %v", err)
		}
		posts = append(posts, product)
	}
	return posts, nil
}

func GetProductByID(id int) (*Product, error) {
	var product Product
	err := DB.QueryRow("SELECT id, model, company, price FROM products WHERE id = $1", id).
		Scan(&product.ID, &product.Model, &product.Company, &product.Price)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("продукт с ID %d не найден", id)
		}
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	return &product, nil
}

func AddProduct(model, company string, price int) error {
	_, err := DB.Exec(
		"INSERT INTO products (model, company, price) VALUES ($1, $2, $3)",
		model, company, price,
	)
	if err != nil {
		return fmt.Errorf("ошибка добавления продукта: %v", err)
	}
	return nil
}

func UpdateProduct(id int, model, company string, price int) error {
	_, err := DB.Exec(
		"UPDATE products SET model=$1, company=$2, price=$3 WHERE id=$4",
		model, company, price, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления продукта: %v", err)
	}
	return nil
}

func DeleteProduct(id int) error {
	_, err := DB.Exec("DELETE FROM products WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления продукта: %v", err)
	}
	return nil
}
