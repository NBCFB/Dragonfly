package auth

import (
	"log"
	"net/http"
	"time"
)

var AuthMiddleware = func(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cost int64
		now := time.Now()
		tokenString := r.Header.Get("Authorization")

		isTokenValid, err := ValidateToken(tokenString)
		if err != nil {
			log.Println(err.Error())
		}

		if isTokenValid && err == nil {
			f(w, r)
		} else {
			log.Println(err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(""))
		}

		cost = time.Since(now).Nanoseconds() / int64(time.Millisecond)
		log.Println(cost)
		// Log API request for later mining
	}
}
