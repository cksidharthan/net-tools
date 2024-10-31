package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/cksidharthan/net-tools/pkg"
)

func main() {
	chiRouter := chi.NewRouter()
	chiRouter.Use(middleware.Logger)
	chiRouter.Use(middleware.Recoverer)
	chiRouter.Use(middleware.URLFormat)

	chiRouter.Get("/ping", pkg.PingHandler)

	http.ListenAndServe(":3000", chiRouter)
}
