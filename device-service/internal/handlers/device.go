package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/google/uuid"

    "github.com/elenaniknovikova/architecture-pro-warmhouse/device-service/internal/models"
    "github.com/elenaniknovikova/architecture-pro-warmhouse/device-service/internal/repository"
)

type DeviceHandler struct {
    repo repository.DeviceRepository
}

func NewDeviceHandler(repo repository.DeviceRepository) *DeviceHandler {
    return &DeviceHandler{repo: repo}
}

func (h *DeviceHandler) CreateDevice(w http.ResponseWriter, r *http.Request) {
    var req models.CreateDeviceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    device := &models.Device{
        ID:        uuid.New().String(),
        Name:      req.Name,
        Type:      req.Type,
        Room:      req.Room,
        Status:    "offline",
        UserID:    req.UserID,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    if err := h.repo.Create(device); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
    devices, err := h.repo.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(devices)
}

func (h *DeviceHandler) GetDeviceByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    device, err := h.repo.GetByID(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    device, err := h.repo.GetByID(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    var req models.UpdateDeviceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if req.Name != "" {
        device.Name = req.Name
    }
    if req.Room != "" {
        device.Room = req.Room
    }
    if req.Status != "" {
        device.Status = req.Status
    }
    device.UpdatedAt = time.Now()

    if err := h.repo.Update(device); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(device)
}

func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    if err := h.repo.Delete(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}