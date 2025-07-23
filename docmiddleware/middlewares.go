package middlewares

import (
    "encoding/json"
    "net/http"

    "github.com/golang-jwt/jwt"
)

// DoctorOnly ensures the request has a valid JWT with role "doctor" in its claims.
func DoctorOnly(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Retrieve claims from context (populated by VerifyTokenMiddleware)
        claims, ok := r.Context().Value("claims").(jwt.MapClaims)
        if !ok || claims == nil {
            jsonError(w, "Access Denied. No valid token found", http.StatusUnauthorized)
            return
        }

        // Check the "role" claim
        roleValue, exists := claims["role"]
        role, isString := roleValue.(string)
        if !exists || !isString || role != "doctor" {
            jsonError(w, "Access Denied. Doctor privileges required", http.StatusForbidden)
            return
        }

        // Proceed to the next handler
        next.ServeHTTP(w, r)
    })
}

// jsonError writes a JSON error response
func jsonError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{"message": message})
}
