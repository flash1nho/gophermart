package helpers

import (
    "net/http"
    "context"

    "github.com/flash1nho/go-musthave-diploma-tpl/middlewares"
)

func SetSignedCookie(userID int, w http.ResponseWriter) error {
    encodedValue, err := middlewares.SecureCookieManager.Encode(middlewares.CookieName, userID)

    if err != nil {
        return err
    }

    cookie := &http.Cookie{
        Name:     middlewares.CookieName,
        Value:    encodedValue,
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        MaxAge:   3600 * 24 * 7,
    }

    http.SetCookie(w, cookie)

    return nil
}

func GetUserIDFromContext(ctx context.Context) int {
    userID, _ := ctx.Value(middlewares.CtxUserKey).(int)

    return userID
}
