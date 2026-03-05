# Сценарий 3: Получение текущей температуры с датчика

## 1. Участники

- **Клиент**: мобильное приложение пользователя / веб-интерфейс
- **Device Service**: сервис управления устройствами
- **Telemetry Service**: сервис сбора и хранения телеметрии
- **Integration Service**: сервис интеграции с реальными устройствами
- **Auth Service**: сервис аутентификации

## 2. Описание

Пользователь хочет узнать текущую температуру в комнате. Приложение запрашивает последние показания с датчика. Если данные есть в базе телеметрии (не старше 1 минуты), они возвращаются сразу. Если данных нет или они устарели, Device Service может запросить актуальные показания через Integration Service.

## 3. Последовательность шагов

1. Пользователь открывает экран "Климат" в приложении
2. Приложение отправляет GET-запрос в Device Service для получения списка датчиков
3. Для каждого датчика Device Service запрашивает последние показания у Telemetry Service
4. Telemetry Service возвращает сохранённые данные
5. Если данные отсутствуют или устарели (опционально), Device Service может запросить актуальные через Integration Service
6. Device Service возвращает пользователю список датчиков с текущими показаниями

## 4. Детали запроса от клиента

### 4.1 Получение списка датчиков с показаниями

**Endpoint:** `GET /api/v1/devices?type=temperature`

### Заголовки (Headers)

|    Header     |       Значение    | Обязательный |        Описание         |
|---------------|-------------------|--------------|-------------------------|
| Authorization | `Bearer {token}`  |      Да      | JWT токен пользователя  |
| Accept        | `application/json`|      Нет     | Ожидаемый формат ответа |

### Параметры запроса (Query Parameters)

|      Параметр       |    Тип  | Обязательный |          Описание          |     Пример    |
|---------------------|---------|--------------|----------------------------|---------------|
| `type`              | string  |     Нет      | Фильтр по типу устройства  | `temperature` |
| `room`              | string  |     Нет      | Фильтр по комнате          | `living_room` |
| `include_telemetry` | boolean |     Нет      | Включить текущие показания | `true`        |

### Пример полного запроса

```http
GET /api/v1/devices?type=temperature&room=living_room&include_telemetry=true HTTP/1.1
Host: api.smarthome.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9...
Accept: application/json

```

### Ответ при успехе (200 OK)

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "devices": [
    {
      "id": "temp-123",
      "name": "Living Room Sensor",
      "type": "temperature",
      "room": "living_room",
      "status": "online",
      "telemetry": {
        "temperature": 23.5,
        "humidity": 45,
        "battery": 89,
        "timestamp": "2026-02-20T15:30:00Z",
        "age_seconds": 5
      }
    },
    {
      "id": "temp-456",
      "name": "Kitchen Sensor",
      "type": "temperature",
      "room": "kitchen",
      "status": "online",
      "telemetry": {
        "temperature": 24.2,
        "humidity": 52,
        "battery": 76,
        "timestamp": "2026-02-20T15:29:30Z",
        "age_seconds": 35
      }
    }
  ],
  "total": 2,
  "retrieved_at": "2026-02-20T15:30:05Z"
}
```
**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1devices/get)
- **gRPC:** [`telemetry.TelemetryService/GetLatestTelemetry`](../proto/telemetry.proto)
- **gRPC:** [`integration.IntegrationService/ReadTelemetry`](../proto/integration.proto)
- **MQTT:** [`asyncapi.yaml`](../asyncapi/asyncapi.yaml#/channels/devices~1{deviceId}~1get)


### 4.2 Получение показаний конкретного датчика

**Endpoint:** `GET /api/v1/telemetry/devices/{deviceId}/current`

### Параметры пути (Path Parameters)

|  Параметр |   Тип  | Обязательный |  Описание  |
|-----------|--------|--------------|------------|
| `deviceId`| string |      Да      | ID датчика |

### Пример запроса

```http
GET /api/v1/telemetry/devices/temp-123/current HTTP/1.1
Host: api.smarthome.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9...
```

### Ответ

```json
{
  "deviceId": "temp-123",
  "deviceName": "Living Room Sensor",
  "metrics": {
    "temperature": 23.5,
    "humidity": 45,
    "battery": 89,
    "signal_strength": -67
  },
  "timestamp": "2026-02-20T15:30:00Z",
  "age_seconds": 5,
  "source": "cache"
}
```
**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1telemetry~1devices~1{deviceId}~1current/get)
- **gRPC:** [`telemetry.TelemetryService/GetLatestTelemetry`](../proto/telemetry.proto)


## 5. Межсервисное взаимодействие

### 5.1 Device Service → Telemetry Service (запрос последних показаний)

**Сервис:** `telemetry.TelemetryService`
**Метод:** `GetLatestTelemetry`
**Протокол:** gRPC

**Request (GetLatestRequest):**
```protobuf
message GetLatestRequest {
  string device_id = 1;           // ID устройства
  repeated string metrics = 2;     // список метрик (опционально)
  int32 max_age_seconds = 3;       // максимальный возраст данных
}
```

**Пример запроса:**
```json
{
  "device_id": "temp-123",
  "metrics": ["temperature", "humidity"],
  "max_age_seconds": 60
}
```

**Response (GetLatestResponse):**
```protobuf
message GetLatestResponse {
  string device_id = 1;
  TelemetryPoint data = 2;
  bool from_cache = 3;
}

message TelemetryPoint {
  google.protobuf.Timestamp timestamp = 1;
  map<string, double> numeric_metrics = 2;
  map<string, string> string_metrics = 3;
}
```

**Пример ответа:**
```json
{
  "device_id": "temp-123",
  "data": {
    "timestamp": "2026-02-20T15:30:00Z",
    "numeric_metrics": {
      "temperature": 23.5,
      "humidity": 45
    }
  },
  "from_cache": true
}
```

**Описание полей:**

|        Поле           |   Тип    |           Описание                  |
|-----------------------|----------|-------------------------------------|
| `device_id`           | string   | ID устройства                       |
| `data.timestamp`      | datetime | Время измерения                     |
| `data.numeric_metrics`| map      | Числовые показатели                 |
| `data.string_metrics` | map      | Строковые показатели                |
| `from_cache`          | boolean  | Данные из кэша или реального опроса |

### 5.2 Device Service → Integration Service (запрос актуальных данных)

Если данных нет в Telemetry Service или они устарели, Device Service может запросить актуальные показания напрямую через Integration Service.

**Сервис:** `integration.IntegrationService`
**Метод:** `ReadTelemetry`
**Протокол:** gRPC

**Request (ReadTelemetryRequest):**
```protobuf
message ReadTelemetryRequest {
  string device_id = 1;
  repeated string metrics = 2;
  int32 timeout_ms = 3;
}
```

**Пример запроса:**
```json
{
  "device_id": "temp-123",
  "metrics": ["temperature", "humidity"],
  "timeout_ms": 5000
}
```

**Response (ReadTelemetryResponse):**
```protobuf
message ReadTelemetryResponse {
  bool success = 1;
  map<string, double> values = 2;
  google.protobuf.Timestamp timestamp = 3;
  int32 response_time_ms = 4;
  string error = 5;
}
```

**Пример ответа:**
```json
{
  "success": true,
  "values": {
    "temperature": 23.5,
    "humidity": 45
  },
  "timestamp": "2026-02-20T15:30:02Z",
  "response_time_ms": 150
}
```

### 5.3 Integration Service → MQTT (опрос устройства)

**Топик:** `devices/{deviceId}/get`
**Протокол:** MQTT

**Сообщение:**
```json
{
  "command": "read",
  "metrics": ["temperature", "humidity"],
  "request_id": "req-789",
  "timestamp": "2026-02-20T15:30:02Z"
}
```

**Ответ от устройства (в топик `devices/{deviceId}/telemetry`):**
```json
{
  "request_id": "req-789",
  "temperature": 23.5,
  "humidity": 45,
  "battery": 89,
  "timestamp": "2026-02-20T15:30:03Z"
}
```

## 6. Кэширование данных

Для оптимизации производительности используется многоуровневое кэширование:

### Уровень 1: Telemetry Service (InfluxDB)
- Хранит исторические данные
- Retention policy: 30 дней для сырых данных, 1 год для агрегированных

### Уровень 2: Redis (кэш последних значений)
- Хранит последние показания каждого датчика
- TTL: 60 секунд
- Ключ: `telemetry:latest:{device_id}`

### Уровень 3: Device Service (in-memory кэш)
- Кэширует ответы для часто запрашиваемых устройств
- TTL: 5 секунд
- Используется для уменьшения нагрузки на Telemetry Service

```python
# Псевдокод логики получения данных
def get_device_telemetry(device_id):
    # Проверка in-memory кэша
    cached = memory_cache.get(device_id)
    if cached and cached.age < 5:
        return cached
    
    # Запрос в Redis (последние значения)
    redis_data = redis.get(f"telemetry:latest:{device_id}")
    if redis_data and redis_data.age < 60:
        memory_cache.set(device_id, redis_data, ttl=5)
        return redis_data
    
    # Запрос в Telemetry Service
    telemetry = telemetry_client.GetLatestTelemetry(device_id)
    
    # Сохранение в кэши
    redis.setex(f"telemetry:latest:{device_id}", 60, telemetry)
    memory_cache.set(device_id, telemetry, ttl=5)
    
    return telemetry
```

## 7. Обработка ошибок

| HTTP код | Описание                     |                    Пример ответа                        |
|----------|------------------------------|---------------------------------------------------------|
|   400    | Неверный формат запроса      | `{"error": "Invalid device ID format"}`                 |
|   401    | Не авторизован               | `{"error": "Token expired"}`                            |
|   403    | Нет прав                     | `{"error": "User does not have access to this device"}` |
|   404    | Устройство не найдено        | `{"error": "Device not found"}`                         |
|   408    | Таймаут запроса к устройству | `{"error": "Device did not respond in time"}`           |
|   423    | Устройство заблокировано     | `{"error": "Device is busy, try again later"}`          |
|   503    | Сервис недоступен            | `{"error": "Telemetry service unavailable"}`            |
|   504    | Шлюз недоступен              | `{"error": "Integration service timeout"}`              |

## 8. Диаграмма последовательности (Mermaid)

```mermaid
sequenceDiagram
    participant Client as Клиент (App)
    participant Device as Device Service
    participant Telemetry as Telemetry Service
    participant Redis as Redis Cache
    participant Integration as Integration Service
    participant MQTT as MQTT Broker

    Client->>Device: GET /devices?type=temperature
    
    activate Device
    Device->>Device: Проверка in-memory кэша
    Note right of Device: Кэш пуст или устарел
    
    Device->>Redis: GET telemetry:latest:temp-123
    activate Redis
    Redis-->>Device: Данные есть (age 30s)
    deactivate Redis
    
    Device->>Telemetry: GetLatestTelemetry(temp-123)
    activate Telemetry
    
    Telemetry->>Telemetry: Проверка InfluxDB
    Note right of Telemetry: Данные есть (age 45s)
    Telemetry-->>Device: {temperature: 23.5, timestamp: ...}
    deactivate Telemetry
    
    Device-->>Client: Ответ с данными
    
    Client->>Device: GET /telemetry/devices/temp-124/current
    activate Device
    
    Device->>Redis: GET telemetry:latest:temp-124
    activate Redis
    Redis-->>Device: Данных нет
    deactivate Redis
    
    Device->>Telemetry: GetLatestTelemetry(temp-124)
    activate Telemetry
    Telemetry-->>Device: Данных нет (никогда не было)
    deactivate Telemetry
    
    Device->>Integration: ReadTelemetry(temp-124)
    activate Integration
    
    Integration->>MQTT: publish get command
    activate MQTT
    MQTT-->>Integration: command sent
    deactivate MQTT
    
    Integration-->>Device: Запрос отправлен, ожидание...
    
    MQTT->>Integration: device telemetry received
    activate Integration
    Integration-->>Device: {temperature: 22.8, timestamp: ...}
    deactivate Integration
    
    Device->>Redis: SETEX telemetry:latest:temp-124 (60s)
    activate Redis
    Redis-->>Device: OK
    deactivate Redis
    
    Device-->>Client: Ответ с актуальными данными
    deactivate Device
```

## 9. Примеры запросов для разных сценариев

### Сценарий А: Данные есть в кэше (быстрый ответ)

```http
GET /api/v1/telemetry/devices/temp-123/current
```

```json
{
  "deviceId": "temp-123",
  "temperature": 23.5,
  "humidity": 45,
  "timestamp": "2026-02-20T15:30:00Z",
  "age_seconds": 3,
  "source": "cache"
}
```

### Сценарий Б: Данных нет, опрос устройства (медленный ответ)

```http
GET /api/v1/telemetry/devices/temp-456/current?force_refresh=true
```
```json
{
  "deviceId": "temp-456",
  "temperature": 22.1,
  "humidity": 48,
  "timestamp": "2026-02-20T15:31:00Z",
  "age_seconds": 0,
  "source": "live",
  "response_time_ms": 1200
}
```

### Сценарий В: Массовое получение данных для комнаты

```http
GET /api/v1/rooms/living_room/telemetry
```
```json
{
  "room": "living_room",
  "devices": [
    {
      "id": "temp-123",
      "name": "Temperature Sensor",
      "temperature": 23.5
    },
    {
      "id": "hum-123",
      "name": "Humidity Sensor",
      "humidity": 45
    }
  ],
  "average_temperature": 23.5,
  "average_humidity": 45,
  "timestamp": "2026-02-20T15:30:00Z"
}
```

## 10. Метрики и мониторинг

Telemetry Service собирает метрики о своей работе:

|            Метрика             |           Описание          | Где смотреть |
|--------------------------------|-----------------------------|--------------|
| `telemetry.requests.total`     | Всего запросов              | Prometheus   |
| `telemetry.requests.cache_hit` | Попадания в кэш             | Prometheus   |
| `telemetry.requests.live`      | Реальные опросы устройств   | Prometheus   |
| `telemetry.latency_ms`         | Время ответа                | Prometheus   |
| `telemetry.device.offline`     | Количество офлайн устройств | Grafana      |

## 11. Варианты для разных типов устройств

|      Тип устройства     |               Метрики                 |                    Формат ответа                          |
|-------------------------|---------------------------------------|-----------------------------------------------------------|
| Температурный датчик    | temperature, humidity, battery        | `{"temperature": 23.5}`                                   |
| Датчик движения         | motion, battery                       | `{"motion": true, "last_motion": "2026-02-20T15:29:00Z"}` |
| Смарт-розетка           | power, energy, voltage                | `{"power": 120.5, "energy_today": 2.4}`                   |
| Датчик качества воздуха | pm25, pm10, co2, tvoc                 | `{"pm25": 12.3, "co2": 450}`                              |
| Метеостанция            | temperature, humidity, pressure, wind | `{"pressure": 1013.2, "wind_speed": 5.2}`                 |
```
```
## 12. Связанные спецификации

- **REST API:** [`openapi.yaml`](../openapi/openapi.yaml)
- **gRPC API:** [`telemetry.proto`](../proto/telemetry.proto), [`integration.proto`](../proto/integration.proto)
- **AsyncAPI:** [`asyncapi.yaml`](../asyncapi/asyncapi.yaml)
```
