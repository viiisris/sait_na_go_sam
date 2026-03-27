package manager

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Получаем все продукты из БД
	products, err := GetAllProducts()
	if err != nil {
		log.Printf("Ошибка получения продуктов: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ht, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		log.Printf("Ошибка загрузки файла index: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Передаем продукты в шаблон
	err = ht.Execute(w, products)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AddProductForm(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что это GET запрос
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ht, err := template.ParseFiles("./templates/add_product.html")
	if err != nil {
		log.Printf("Ошибка загрузки файла add_product: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = ht.Execute(w, nil)
	if err != nil {
		log.Printf("Ошибка выполнения шаблона: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func AddProductSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	err := r.ParseForm()
	if err != nil {
		log.Printf("Ошибка парсинга формы: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	model := r.FormValue("model")
	company := r.FormValue("company")
	priceStr := r.FormValue("price")

	// Валидация данных
	if model == "" || company == "" || priceStr == "" {
		http.Error(w, "Все поля обязательны для заполнения", http.StatusBadRequest)
		return
	}

	// Преобразуем цену в число
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		http.Error(w, "Цена должна быть числом", http.StatusBadRequest)
		return
	}

	err = AddProduct(model, company, price)
	if err != nil {
		log.Printf("Ошибка добавления товара: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func EditProductForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/edit-product/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		product, err := GetProductByID(id)
		if err != nil {
			log.Printf("Ошибка получения продукта: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		ht, err := template.ParseFiles("./templates/edit_product.html")
		if err != nil {
			log.Printf("Ошибка загрузки файла edit_product: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ht, err = template.ParseFiles("./templates/edit_product.html")
		if err != nil {
			log.Printf("Ошибка загрузки файла edit_product: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		err = ht.Execute(w, product)
		if err != nil {
			log.Printf("Ошибка выполнения шаблона: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		// Получаем данные из формы
		err := r.ParseForm()
		if err != nil {
			log.Printf("Ошибка парсинга формы: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		model := r.FormValue("model")
		company := r.FormValue("company")
		priceStr := r.FormValue("price")

		if model == "" || company == "" || priceStr == "" {
			http.Error(w, "Все поля обязательны для заполнения", http.StatusBadRequest)
			return
		}

		price, err := strconv.Atoi(priceStr)
		if err != nil {
			http.Error(w, "Цена должна быть числом", http.StatusBadRequest)
			return
		}

		err = UpdateProduct(id, model, company, price)
		if err != nil {
			log.Printf("Ошибка обновления товара: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func DeleteProductForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/delete-product/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = DeleteProduct(id)
	if err != nil {
		log.Printf("Ошибка удаления товара: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
