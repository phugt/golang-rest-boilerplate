package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anyshare/anyshare-admin-api/api"
	"github.com/anyshare/anyshare-admin-api/middlewares"
	"github.com/anyshare/anyshare-common/mongodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	if os.Getenv("PORT") == "" {
		panic("Cannot load app configuration, exit app!")
	}

	mongodb.Connect()
	r := initRouter()
	setupAPI(r)
	startServer(r)
}

func initRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(middlewares.LocaleHeader)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	return r
}

func setupAPI(r *chi.Mux) {
	//Public
	r.Group(func(r chi.Router) {
		r.Post("/login", api.Login)
	})
	//Protected
	r.Group(func(r chi.Router) {
		r.Use(middlewares.Authentication)
		r.Get("/profile", api.GetProfile)
		r.Post("/profile", api.UpdateProfile)
		r.Post("/profile/password", api.ChangePassword)

		r.Get("/user", api.ListUser)
		r.Get("/user/{id}", api.GetUser)
		r.Post("/user", api.CreateUser)
		r.Put("/user", api.UpdateUser)
		r.Delete("/user/{id}", api.DeleteUser)

		r.Get("/admin", api.ListAdmin)
		r.Get("/admin/{id}", api.GetAdmin)
		r.Post("/admin", api.GetAdmin)
		r.Put("/admin", api.UpdateAdmin)
		r.Delete("/admin/{id}", api.DeleteAdmin)
	})
}

func startServer(r *chi.Mux) {
	server := &http.Server{Addr: ":" + os.Getenv("PORT"), Handler: r}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		mongodb.Disconnect()
		shutdownCtx, shutdownCtxCancel := context.WithTimeout(serverCtx, 30*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()
		shutdownCtxCancel()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-serverCtx.Done()
}
