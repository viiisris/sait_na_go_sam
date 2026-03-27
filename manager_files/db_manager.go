package manager

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Product struct {
	ID      int
	Model   string
	Company string
	Price   int
}

// Глобальная переменная для доступа к БД из других функций
var DB *sql.DB

// Функция для инициализации подключения к БД
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

// Функция для закрытия подключения к БД
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// Функция для получения всех продуктов
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	return posts, nil
}

// Функция для получения продукта по ID
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

// Добавление нового продукта
func AddProduct(model, company string, price int) error {
	// Используем параметризованный запрос для защиты от SQL-инъекций
	_, err := DB.Exec(
		"INSERT INTO products (model, company, price) VALUES ($1, $2, $3)",
		model, company, price,
	)
	if err != nil {
		return fmt.Errorf("ошибка добавления продукта: %v", err)
	}
	return nil
}

// Редактирование существующего продукта
func UpdateProduct(id int, model, company string, price int) error {
	_, err := DB.Exec(
		"UPDATE products SET model = $1, company = $2, price = $3 WHERE id = $4",
		model, company, price, id,
	)
	if err != nil {
		return fmt.Errorf("ошибка редактирования продукта: %v", err)
	}
	return nil
}

func DeleteProduct(id int) error {
	_, err := DB.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("ошибка удаления продукта: %v", err)
	}
	return nil
}
