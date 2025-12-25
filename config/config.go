package config

import (
    "flag"
    "errors"
    "strings"
    "strconv"
    "os"

    "github.com/flash1nho/go-musthave-diploma-tpl/logger"

    "go.uber.org/zap"
)

const (
    DefaultHost = "localhost:8080"
    DefaultURL = "http://localhost:8080"
)

type Server struct {
    Addr string
    BaseURL string
}

type NetAddress struct {
    Host string
    Port int
}

func (addr *NetAddress) String() string {
    return addr.Host + ":" + strconv.Itoa(addr.Port)
}

func (addr *NetAddress) Set(s string) error {
    trimmed := strings.TrimPrefix(s, "http://")
    hp := strings.Split(trimmed, ":")

    if len(hp) != 2 {
        return errors.New("значение может быть таким: " + DefaultHost + "|" + DefaultURL)
    }

    port, err := strconv.Atoi(hp[1])

    if err != nil {
        return err
    }

    addr.Host = hp[0]
    addr.Port = port

    return nil
}

func Settings() (Server, Server, *zap.Logger, string) {
    apiAddress := new(NetAddress).String()
    flag.StringVar(&apiAddress, "a", DefaultHost, "реквизиты API сервиса")

    accrualAddress := new(NetAddress).String()
    flag.StringVar(&accrualAddress, "r", DefaultHost, "реквизиты Accrual сервиса")

    var databaseURI string
    flag.StringVar(&databaseURI, "d", "", "реквизиты базы данных")

    flag.Parse()

    envDatabaseURI, ok := os.LookupEnv("DATABASE_URI")

    if ok {
        databaseURI = envDatabaseURI
    }

    logger.Initialize("info")

    return ServerData(apiAddress, "RUN_ADDRESS"),
           ServerData(accrualAddress, "ACCRUAL_SYSTEM_ADDRESS"),
           logger.Log,
           databaseURI
}

func ServerData(serverAddress string, envName string) Server {
    if envServerAddress := os.Getenv(envName); envServerAddress != "" {
        serverAddress = envServerAddress
    } else if serverAddress == ":0" {
        serverAddress = DefaultHost
    }

    trimmedServerAddress := strings.TrimPrefix(serverAddress, "http://")
    serverBaseURL := "http://" + trimmedServerAddress

    if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
        serverBaseURL = envBaseURL
    }

    return Server{Addr: serverAddress, BaseURL: serverBaseURL}
}
