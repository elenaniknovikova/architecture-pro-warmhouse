package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"

    "github.com/elenaniknovikova/architecture-pro-warmhouse/device-service/internal/handlers"
    "github.com/elenaniknovikova/architecture-pro-warmhouse/device-service/internal/repository"
)

func main() {
    // Инициализируем репозиторий (in-memory)
    repo := repository.NewInMemoryRepository()
    
    // Инициализируем обработчики
    deviceHandler := handlers.NewDeviceHandler(repo)

    // Создаём роутер
    r := mux.NewRouter()

    // Health check
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    // API routes
    api := r.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/devices", deviceHandler.CreateDevice).Methods("POST")
    api.HandleFunc("/devices", deviceHandler.GetDevices).Methods("GET")
    api.HandleFunc("/devices/{id}", deviceHandler.GetDeviceByID).Methods("GET")
    api.HandleFunc("/devices/{id}", deviceHandler.UpdateDevice).Methods("PUT")
    api.HandleFunc("/devices/{id}", deviceHandler.DeleteDevice).Methods("DELETE")

    // Определяем порт
    port := os.Getenv("PORT")
    if port == "" {
        port = "8082"
    }

    log.Printf("Device Service is running on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}