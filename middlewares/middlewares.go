package middlewares

import (
	  "compress/gzip"
	  "net/http"
	  "context"

    "github.com/gorilla/securecookie"
)

var hashKey = securecookie.GenerateRandomKey(32)
var SecureCookieManager = securecookie.New(hashKey, nil)

const CookieName = "user_session_id"

type ctxUserID string
const CtxUserKey = ctxUserID("userID")

func Decompressor(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Header.Get("Content-Encoding") == "gzip" {
            gzReader, err := gzip.NewReader(r.Body)
            if err != nil {
                http.Error(w, "ошибка при распаковке gzip", http.StatusBadRequest)
                return
            }

            defer gzReader.Close()

            r.Body = gzReader
        }

        next.ServeHTTP(w, r)
    })
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie(CookieName)

		if err != nil {
				http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
				return
		}

    var userID int

		err = SecureCookieManager.Decode(CookieName, cookie.Value, &userID)

		if err != nil {
				http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized)
				return
		}

		ctx := context.WithValue(r.Context(), CtxUserKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
