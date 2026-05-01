package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/platonso/hrmate-api/docs"
	"github.com/platonso/hrmate-api/internal/config"
	"github.com/platonso/hrmate-api/internal/domain"
	"github.com/platonso/hrmate-api/internal/handler/auth"
	"github.com/platonso/hrmate-api/internal/handler/form"
	"github.com/platonso/hrmate-api/internal/handler/middleware"
	"github.com/platonso/hrmate-api/internal/handler/user"
)

type AuthProvider interface {
	auth.Service
	middleware.AuthService
}

type UserProvider interface {
	user.Service
	middleware.UserService
}

type Router struct {
	handlerAuth *auth.Handler
	handlerUser *user.Handler
	handlerForm *form.Handler
	middleware  *middleware.Auth
	cfg         *config.Config
}

func NewRouter(cfg *config.Config, authSvc AuthProvider, userSvc UserProvider, formSvc form.Service,
) *Router {
	authMiddleware := &middleware.Auth{
		AuthSvc: authSvc,
		UserSvc: userSvc,
	}

	return &Router{
		handlerAuth: auth.NewHandler(authSvc),
		handlerUser: user.NewHandler(userSvc),
		handlerForm: form.NewHandler(formSvc),
		middleware:  authMiddleware,
		cfg:         cfg,
	}
}

func (rt *Router) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Timeout(14 * time.Second))

	// Authentication
	r.Post("/register", rt.handlerAuth.HandleRegister)
	r.Post("/login", rt.handlerAuth.HandleLogin)

	r.Group(func(r chi.Router) {
		r.Use(rt.middleware.AuthMiddleware)
		r.Use(rt.middleware.RequireActiveStatus)

		r.Get("/me", rt.handlerUser.HandleGetMe)
	})

	// Employee
	r.Group(func(r chi.Router) {
		r.Use(rt.middleware.AuthMiddleware)
		r.Use(rt.middleware.RequireRoles(domain.RoleEmployee))
		r.Use(rt.middleware.RequireActiveStatus)

		r.Post("/forms", rt.handlerForm.HandleCreateForm)
		r.Get("/forms", rt.handlerForm.HandleGetForms)
		r.Get("/forms/{id}", rt.handlerForm.HandleGetForm)
	})

	// HR
	r.Route("/hr", func(r chi.Router) {
		r.Use(rt.middleware.AuthMiddleware)
		r.Use(rt.middleware.RequireRoles(domain.RoleHR))
		r.Use(rt.middleware.RequireActiveStatus)

		r.Get("/user/{id}", rt.handlerUser.HandleGetUser)
		r.Get("/users", rt.handlerUser.HandleGetUsers)

		r.Get("/forms", rt.handlerForm.HandleGetForms)

		r.Get("/forms/{id}", rt.handlerForm.HandleGetForm)
		r.Patch("/forms/{id}/approve", rt.handlerForm.HandleApprove)
		r.Patch("/forms/{id}/reject", rt.handlerForm.HandleReject)
	})

	// Admin
	r.Route("/admin", func(r chi.Router) {
		r.Use(rt.middleware.AuthMiddleware)
		r.Use(rt.middleware.RequireRoles(domain.RoleAdmin))
		r.Use(rt.middleware.RequireActiveStatus)

		r.Get("/user/{id}", rt.handlerUser.HandleGetUser)
		r.Get("/users", rt.handlerUser.HandleGetUsers)
		r.Patch("/users/{id}/activate", rt.handlerUser.HandleActivate)
		r.Patch("/users/{id}/deactivate", rt.handlerUser.HandleDeactivate)

		r.Get("/forms", rt.handlerForm.HandleGetForms)

		r.Get("/forms/{id}", rt.handlerForm.HandleGetForm)
		//r.Patch("/forms/{id}/approve", rt.handlerForm.HandleApprove)
		//r.Patch("/forms/{id}/reject", rt.handlerForm.HandleReject)
		r.Delete("/forms/{id}", rt.handlerForm.HandleDelete)
	})

	// Documents
	r.Route("/documents", func(r chi.Router) {
		r.Use(rt.middleware.AuthMiddleware)
		r.Use(rt.middleware.RequireActiveStatus)

		r.Get("/{id}/download", rt.handlerForm.HandleDownloadDocument)
	})

	return r
}
