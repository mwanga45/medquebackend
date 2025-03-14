package api

import "net/http"

func doctors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet{
		http.Error(w,"Invalid methode used", http.StatusBadRequest)
      return
	}
	query := "SELECT * FROM doctors"

}