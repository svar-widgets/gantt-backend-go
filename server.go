package main

import (
	"gantt-backend-go/data"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/unrolled/render"
)

var format = render.New()

// Config is the structure that stores the settings for this backend app
var Config AppConfig

func main() {
	readConfig()

	dao := data.NewDAO(Config.DB, Config.Server.URL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if len(Config.Server.Cors) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins:   Config.Server.Cors,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
			AllowCredentials: true,
			MaxAge:           300,
		})
		r.Use(c.Handler)
	}

	initRoutes(r, dao)

	log.Printf("Starting webserver at port " + Config.Server.Port)
	err := http.ListenAndServe(Config.Server.Port, r)
	if err != nil {
		log.Println(err.Error())
	}
}

func readConfig() {
	k := koanf.New(".")
	f := file.Provider("config.yml")
	if err := k.Load(f, yaml.Parser()); err != nil {
		log.Println("warning, can't load config from config.yml")
	}

	k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "APP_")), "_", ".", -1)
	}), nil)

	k.Unmarshal("", &Config)
}
