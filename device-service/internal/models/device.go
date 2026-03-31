package models

import "time"

type Device struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Type      string    `json:"type"`       // "light", "thermostat", "sensor"
    Room      string    `json:"room"`
    Status    string    `json:"status"`     // "online", "offline"
    UserID    string    `json:"user_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type CreateDeviceRequest struct {
    Name   string `json:"name"`
    Type   string `json:"type"`
    Room   string `json:"room"`
    UserID string `json:"user_id"`
}

type UpdateDeviceRequest struct {
    Name   string `json:"name"`
    Room   string `json:"room"`
    Status string `json:"status"`
}
