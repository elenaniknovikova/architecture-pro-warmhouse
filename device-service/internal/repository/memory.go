package repository

import (
    "fmt"
    "sync"


    "github.com/elenaniknovikova/architecture-pro-warmhouse/device-service/internal/models"
)

type DeviceRepository interface {
    Create(device *models.Device) error
    GetByID(id string) (*models.Device, error)
    GetAll() ([]*models.Device, error)
    Update(device *models.Device) error
    Delete(id string) error
}

type InMemoryRepository struct {
    mu      sync.RWMutex
    devices map[string]*models.Device
}

func NewInMemoryRepository() *InMemoryRepository {
    return &InMemoryRepository{
        devices: make(map[string]*models.Device),
    }
}

func (r *InMemoryRepository) Create(device *models.Device) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.devices[device.ID]; exists {
        return fmt.Errorf("device with ID %s already exists", device.ID)
    }
    
    r.devices[device.ID] = device
    return nil
}

func (r *InMemoryRepository) GetByID(id string) (*models.Device, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    device, exists := r.devices[id]
    if !exists {
        return nil, fmt.Errorf("device with ID %s not found", id)
    }
    return device, nil
}

func (r *InMemoryRepository) GetAll() ([]*models.Device, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    devices := make([]*models.Device, 0, len(r.devices))
    for _, device := range r.devices {
        devices = append(devices, device)
    }
    return devices, nil
}

func (r *InMemoryRepository) Update(device *models.Device) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.devices[device.ID]; !exists {
        return fmt.Errorf("device with ID %s not found", device.ID)
    }
    
    r.devices[device.ID] = device
    return nil
}

func (r *InMemoryRepository) Delete(id string) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    if _, exists := r.devices[id]; !exists {
        return fmt.Errorf("device with ID %s not found", id)
    }
    
    delete(r.devices, id)
    return nil
}