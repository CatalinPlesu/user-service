package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/CatalinPlesu/user-service/handler"
	"github.com/CatalinPlesu/user-service/repository/user"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/users", a.loadUserRoutes)

	a.router = router
}

func (a *App) loadUserRoutes(router chi.Router) {
	// userHandler := &handler.User{
	// 	Repo: &user.RedisRepo{
	// 		Client: a.rdb,
	// 	},
	// }

	userHandler := &handler.User{
		Repo: user.NewPostgresRepo(a.db),
		RabbitMQ: a.rabbitMQ,
	}

	router.Get("/", userHandler.List)
	router.Post("/register", userHandler.Register)
	router.Post("/login", userHandler.Login)
	router.Get("/username/{username}", userHandler.GetByUsername)
	router.Get("/displayname/{displayname}", userHandler.GetByDisplayName)
	router.Get("/{id}", userHandler.GetByID)
	router.Put("/{id}", userHandler.UpdateByID)
	router.Delete("/{id}", userHandler.DeleteByID)
}
