package service

import (
		"context"
		"net/http"
		"os"
		"os/signal"
		"syscall"
		"time"
		"fmt"
		"slices"
		"sync"

    "github.com/flash1nho/go-musthave-diploma-tpl/config"
		"github.com/flash1nho/go-musthave-diploma-tpl/middlewares"
		"github.com/flash1nho/go-musthave-diploma-tpl/handler"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

type Service struct {
	  handler *handler.Handler
    servers []config.Server
}

func NewService(handler *handler.Handler, servers []config.Server) *Service {
    return &Service{
    	  handler: handler,
        servers: servers,
    }
}

func (s *Service) mainRouter() http.Handler {
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middlewares.Decompressor)

    r.Post("/api/user/register", s.handler.UserController.Register)
    r.Post("/api/user/login", s.handler.UserController.Login)

    r.Group(func(r chi.Router) {
      	r.Use(middlewares.Auth)
      	r.Post("/api/user/orders", s.handler.OrderController.PostOrders)
      	r.Get("/api/user/orders", s.handler.OrderController.GetOrders)
      	r.Get("/api/user/balance", s.handler.WithdrawnUserController.UserBalance)
      	r.Post("/api/user/balance/withdraw", s.handler.WithdrawnUserController.UserBalanceWithdraw)
      	r.Get("/api/user/withdrawals", s.handler.WithdrawnUserController.UserWithdrawals)
    })

    return r
}

func runServer(s *Service, ctx context.Context, wg *sync.WaitGroup, addr string) {
	defer wg.Done()

	server := &http.Server{
			Addr: addr,
			Handler: s.mainRouter(),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
			s.handler.Log.Info(fmt.Sprintf("Сервер запущен на http://%s", server.Addr))

			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.handler.Log.Error(fmt.Sprintf("ошибка запуска сервера http://%s: %v", server.Addr, err))
			}
	}()

	select {
	case err := <-serverErr:
			s.handler.Log.Info(fmt.Sprint(err))
	case <-ctx.Done():
			s.handler.Log.Info(fmt.Sprintf("Завершение работы сервера http://%s", server.Addr))

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
					s.handler.Log.Info(fmt.Sprintf("Ошибка завершения работы сервера http://%s: %v", server.Addr, err))
			} else {
					s.handler.Log.Info(fmt.Sprintf("Сервер http://%s успешно остановлен", server.Addr))
			}
	}
}

func (s *Service) Run() {
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())

    for _, server := range slices.Compact(s.servers) {
        wg.Add(1)
        go runServer(s, ctx, &wg, server.Addr)
    }

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		sig := <-signalChan
		s.handler.Log.Info(fmt.Sprintf("Полученный сигнал %s: инициирование завершения работы", sig))

		cancel()

		wg.Wait()

		s.handler.Log.Info("Все серверы успешно завершили работу.")
}
