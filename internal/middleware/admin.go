package middleware

import (
	"net/http"

	"github.com/tron-legacy/tron-3d-api/internal/config"
)

// Admin middleware checks if the authenticated user is in the ADMIN_USER_IDS list
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r)
		if userID.IsZero() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userIDHex := userID.Hex()
		isAdmin := false
		for _, adminID := range config.Get().AdminUserIDs {
			if adminID == userIDHex {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
