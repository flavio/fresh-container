package api

import (
	"net/http"
)

func (a *ApiServer) GetHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
