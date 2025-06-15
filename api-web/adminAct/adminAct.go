package adminact

import (
	"encoding/json"
	"net/http"
)
type (
	ServAssignpayload{
		
	}
)

func AssignService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		json.NewEncoder()
	}
}