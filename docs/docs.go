package docs

import (
	"embed"
	"net/http"

	"github.com/gorilla/mux"
)

//go:embed static
var Docs embed.FS

func RegisterOpenAPIService(appName string, rtr *mux.Router) {
	staticServer := http.FileServer(http.FS(Docs))
	rtr.PathPrefix("/static/openapi/").Handler(http.StripPrefix("/static/openapi/", staticServer))
}
