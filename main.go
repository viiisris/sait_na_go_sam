package main

import (
	"log"
	"net/http"
	manager "sait_na_go_sam/manager_files"

	_ "github.com/lib/pq"
)

func main() {

	err := manager.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer manager.CloseDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/register", manager.GuestMiddleware(manager.RegisterPage))
	mux.HandleFunc("/register-submit", manager.GuestMiddleware(manager.Register))
	mux.HandleFunc("/login", manager.GuestMiddleware(manager.LoginPage))
	mux.HandleFunc("/login-submit", manager.GuestMiddleware(manager.Login))

	mux.HandleFunc("/", manager.Home)
	mux.HandleFunc("/add-product", manager.AddProductForm)
	mux.HandleFunc("/add-product-submit", manager.AddProductSubmit)
	mux.HandleFunc("/edit-product/", manager.EditProductForm)
	mux.HandleFunc("/edit-product-submit/", manager.EditProductForm)
	mux.HandleFunc("/delete-product/", manager.DeleteProductForm)

	log.Println("Запуск веб-сервера на http://127.0.0.1:4000")
	err = http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
