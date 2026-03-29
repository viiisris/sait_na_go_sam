package manager

import (
	"context"
	"net/http"
)

type contextKey string

const userContextKey contextKey = "user"

// Middleware для проверки аутентификации
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем cookie с токеном сессии
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// Нет сессии - перенаправляем на страницу входа
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Проверяем сессию
		user, err := GetUserBySessionToken(cookie.Value)
		if err != nil {
			// Сессия недействительна - удаляем cookie и перенаправляем
			http.SetCookie(w, &http.Cookie{
				Name:   "session_token",
				Value:  "",
				MaxAge: -1,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Добавляем пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

// Получение пользователя из контекста
func GetUserFromContext(r *http.Request) *User {
	user, ok := r.Context().Value(userContextKey).(*User)
	if !ok {
		return nil
	}
	return user
}

// Middleware для проверки, что пользователь НЕ аутентифицирован (для страниц логина/регистрации)
func GuestMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil {
			// Проверяем валидность сессии
			user, err := GetUserBySessionToken(cookie.Value)
			if err == nil && user != nil {
				// Уже авторизован - перенаправляем на главную
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
		next(w, r)
	}
}
