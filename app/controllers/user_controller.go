package controllers

import (
    "fmt"
    "net/http"
    "encoding/json"

    "github.com/flash1nho/go-musthave-diploma-tpl/app/models"
    "github.com/flash1nho/go-musthave-diploma-tpl/app/helpers"

    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

type UserController struct {
    Pool *pgxpool.Pool
    Log  *zap.Logger
}

type User struct {
    Login    string `json:"login"`
    Password string `json:"password"`
}

func (controller *UserController) Register(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var user User

    err := json.NewDecoder(r.Body).Decode(&user)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if user.Login == "" || user.Password == "" {
        http.Error(w, "неверный формат запроса", http.StatusBadRequest)
        return
    }

    userID, err := models.UserRegister(r.Context(), user.Login, user.Password, controller.Pool)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if userID == 0 {
        w.WriteHeader(http.StatusConflict)
        fmt.Fprintln(w, "логин уже занят")
        return
    }

    err = helpers.SetSignedCookie(userID, w)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "пользователь успешно зарегистрирован и аутентифицирован")
}

func (controller *UserController) Login(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var user User

    err := json.NewDecoder(r.Body).Decode(&user)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if user.Login == "" || user.Password == "" {
        http.Error(w, "неверный формат запроса", http.StatusBadRequest)
        return
    }

    userID, err := models.UserLogin(r.Context(), user.Login, user.Password, controller.Pool)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    if userID == 0 {
        w.WriteHeader(http.StatusUnauthorized)
        fmt.Fprintln(w, "неверная пара логин/пароль")
        return
    }

    err = helpers.SetSignedCookie(userID, w)

    if err != nil {
        http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
        controller.Log.Error(fmt.Sprint(err))
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "пользователь успешно аутентифицирован")
}
