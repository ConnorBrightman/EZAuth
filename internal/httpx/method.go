package httpx

import "net/http"

func AllowMethod(method string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			Error(
				w,
				http.StatusMethodNotAllowed,
				"method not allowed",
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
