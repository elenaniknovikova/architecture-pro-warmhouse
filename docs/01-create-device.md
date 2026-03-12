# Сценарий 1: Создание нового устройства

## 1. Участники

- **Клиент**: мобильное приложение пользователя
- **Device Service**: сервис управления устройствами
- **Auth Service**: сервис аутентификации
- **Kafka**: шина событий (для асинхронных уведомлений)

## 2. Описание
Пользователь добавляет новую лампочку в свою систему умного дома. 
Приложение отправляет запрос в Device Service, который проверяет права 
пользователя через Auth Service, сохраняет устройство и публикует событие в Kafka.

## 3. Последовательность шагов

1. Пользователь нажимает "Добавить устройство" в приложении
2. Приложение отправляет POST-запрос в Device Service
3. Device Service проверяет токен через Auth Service
4. Auth Service подтверждает, что пользователь существует
5. Device Service сохраняет устройство в базу данных
6. Device Service публикует событие в Kafka (для других сервисов)
7. Device Service возвращает ответ клиенту с ID созданного устройства

## 4. Детали запроса от клиента

Endpoint: `POST /api/v1/devices/`

### Заголовки (Headers)

| Header       | Значение                              | Обязательный | Описание                        |
|--------------|---------------------------------------|--------------|---------------------------------|
| Authorization| `Bearer {token}`                      |      Да      | JWT токен, полученный при входе |
| Content-Type | `application/json`                    |      Да      | Формат данных                   |
| Accept       | `application/json`                    |      Нет     | Ожидаемый формат ответа         |
| X-Request-ID | `550e8400-e29b-41d4-a716-446655440000`|      Нет     | Для отслеживания запросов       |

### Пример полного запроса

```http
POST /api/v1/devices HTTP/1.1
Host: api.smarthome.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyIsIm5hbWUiOiJKb2huIERvZSIsImlhdCI6MTUxNjIzOTAyMn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
Content-Type: application/json
Accept: application/json
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000

```

```json
{
  "name": "Living Room Light",
  "type": "light",
  "room": "living_room"
}

```
**Response (201 Created)**:

```json
{
  "id": "dev-123",
  "name": "Living Room Light",
  "type": "light",
  "room": "living_room",
  "status": "offline",
  "createdAt": "2026-02-19T10:30:00Z"
}

```
**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml)
- **gRPC:** [`auth.AuthService/ValidateToken`](../proto/auth.proto)
- **Kafka:** [`device-events`](../asyncapi/asyncapi.yaml#/channels/device-events)



```http
HTTP/1.1 201 Created
Content-Type: application/json
Location: /api/v1/devices/dev-123
```

## 5. Межсервисное взаимодействие

### 5.1 Device Service → Auth Service (проверка токена)

**Сервис:** `auth.AuthService`
**Метод:** `ValidateToken`
**Протокол:** gRPC

**Request (ValidateTokenRequest):**

```protobuf
message ValidateTokenRequest {
  string token = 1;           // JWT токен для проверки
  string required_permission = 2;  // опционально: требуемое право
}
```
 **Пример запроса**
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9...",
  "required_permission": "devices:create"
}


 **Response (ValidateTokenResponse):**

```protobuf
message ValidateTokenResponse {
  bool valid = 1;              // результат проверки
  string user_id = 2;          // ID пользователя
  string tenant_id = 3;         // ID тенанта (для мультитенантности)
  repeated string roles = 4;    // роли пользователя
  repeated string permissions = 5; // разрешения
  map<string, string> metadata = 6; // дополнительные данные
}
```
 **Пример ответа**

```json
{
  "valid": true,
  "user_id": "user-456",
  "tenant_id": "tenant-789",
  "roles": ["user"],
  "permissions": ["devices:create", "devices:read"],
  "metadata": {
    "email": "user@example.com"
  }
}
```

**Описание полей:**

|     Поле      |   Тип   |                  Описание                     |
|---------------|---------|-----------------------------------------------|
| `valid`       | boolean | `true` если токен валиден                     |
| `user_id`     | string  | Уникальный идентификатор пользователя         |
| `tenant_id`   | string  | Идентификатор клиента (для мультитенантности) |
| `roles`       | array   | Список ролей пользователя                     |
| `permissions` | array   | Список разрешений                             |
| `metadata`    | object  | Дополнительные данные                         |





### 5.2 Device Service - Kafka (публикация события)
Топик: device-events

Сообщение:
```json
{
  "eventId": "evt-123",
  "eventType": "device.created",
  "deviceId": "dev-123",
  "userId": "user-456",
  "timestamp": "2026-02-19T10:30:00Z",
  "data": {
    "name": "Living Room Light",
    "type": "light"
  }
}
```
## 6. Обработка ошибок

|HTTP код |	       Описание	        |Пример ответа                                        |
|---------|-------------------------|-----------------------------------------------------|
|   400	  | Неверный формат запроса	|{"error": "Missing required field: name"}            |
|   401	  | Не авторизован	        |{"error": "Invalid or expired token"}                |
|   403	  | Нет прав	            |{"error": "User does not have permission"}           |
|   409	  | Конфликт	            |{"error": "Device with this name already exists"}    |
|   500	  | Внутренняя ошибка	    |{"error": "Database connection failed"}              |
```