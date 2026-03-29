package manager

import (
	"html/template"
	"log"
	"net/http"
)

// Страница регистрации
func RegisterPage(w http.ResponseWriter, r *http.Request) {
	ht, err := template.ParseFiles("./templates/register.html")
	if err != nil {
		log.Printf("Ошибка загрузки файла register: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ht.Execute(w, nil)
}

// Обработка регистрации
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	// Валидация
	if username == "" || email == "" || password == "" {
		http.Error(w, "Все поля обязательны для заполнения", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		http.Error(w, "Пароли не совпадают", http.StatusBadRequest)
		return
	}

	if len(password) < 6 {
		http.Error(w, "Пароль должен содержать минимум 6 символов", http.StatusBadRequest)
		return
	}

	// Создаем пользователя
	err = CreateUser(username, email, password)
	if err != nil {
		log.Printf("Ошибка регистрации: %v", err)
		http.Error(w, "Пользователь с таким именем или email уже существует", http.StatusBadRequest)
		return
	}

	// Убираем дублирующуюся проверку и лишнюю аутентификацию
	// После успешной регистрации просто перенаправляем на страницу входа
	// Не нужно автоматически логинить пользователя после регистрации

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Страница входа
func LoginPage(w http.ResponseWriter, r *http.Request) {
	ht, err := template.ParseFiles("./templates/login.html")
	if err != nil {
		log.Printf("Ошибка загрузки файла login: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ht.Execute(w, nil)
}

// Обработка входа
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Все поля обязательны для заполнения", http.StatusBadRequest)
		return
	}

	// Аутентификация пользователя
	user, err := AuthenticateUser(username, password)
	if err != nil {
		log.Printf("Ошибка аутентификации: %v", err)
		http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
		return
	}

	// Создаем сессию
	token, err := CreateSession(user.ID)
	if err != nil {
		log.Printf("Ошибка создания сессии: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Устанавливаем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   86400,
	})

	// Перенаправляем на главную
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Выход из системы
func Logout(w http.ResponseWriter, r *http.Request) {
	// Получаем cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Удаляем сессию из БД
		DeleteSession(cookie.Value)

		// Удаляем cookie
		http.SetCookie(w, &http.Cookie{
			Name:   "session_token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	// Перенаправляем на страницу входа
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
