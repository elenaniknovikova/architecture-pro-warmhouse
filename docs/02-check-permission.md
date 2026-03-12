# Сценарий 2: Проверка прав доступа при управлении устройством

## 1. Участники

- **Клиент**: мобильное приложение пользователя
- **Device Service**: сервис управления устройствами
- **Auth Service**: сервис аутентификации
- **Audit Log**: система аудита (для записи действий)

## 2. Описание

Пользователь пытается выполнить действие с устройством (включить/выключить, изменить настройки). Device Service проверяет, имеет ли пользователь право на это действие, запрашивая информацию у Auth Service.

## 3. Последовательность шагов

1. Пользователь нажимает "Включить свет" в приложении
2. Приложение отправляет PATCH-запрос в Device Service
3. Device Service извлекает user_id из JWT токена
4. Device Service запрашивает у Auth Service права пользователя
5. Auth Service возвращает роли и разрешения пользователя
6. Device Service проверяет, есть ли у пользователя право на действие
7. Если право есть — выполняется команда, если нет — возвращается ошибка
8. Device Service записывает действие в аудит-лог

## 4. Детали запроса от клиента

**Endpoint:** `PATCH /api/v1/devices/{deviceId}/state`

### Заголовки (Headers)

| Header        | Значение          | Обязательный | Описание                      |
|---------------|-------------------|--------------|-------------------------------|
| Authorization | `Bearer {token}`  |      Да      | JWT токен пользователя        |
| Content-Type  | `application/json`|      Да      | Формат данных                 |
| If-Match      | `"{etag}"`        |      Нет     | Для предотвращения конфликтов |

### Параметры пути (Path Parameters)

| Параметр   |   Тип  | Обязательный | Описание                         |
|------------|--------|--------------|----------------------------------|
| `deviceId` | string |      Да      | ID устройства, которым управляют |

### Пример полного запроса

```http
PATCH /api/v1/devices/dev-123/state HTTP/1.1
Host: api.smarthome.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9...
Content-Type: application/json
If-Match: "abc123"

```json
{
  "command": "turn_on",
  "parameters": {
    "brightness": 80,
    "transition": 500
  }
}

```
 **Ответ при успехе (200 ОК)**

```http
HTTP/1.1 200 OK
Content-Type: application/json
ETag: "def456"

```json

{
  "deviceId": "dev-123",
  "status": "online",
  "state": {
    "power": "on",
    "brightness": 80
  },
  "lastUpdated": "2026-02-20T14:30:00Z"
}

```

## 5. Межсервисное взаимодействие
### 5.1. Device Service → Auth Service (проверка прав)

**Сервис: auth.AuthService**
**Метод: CheckPermission**
**Протокол: gRPC**

**Request (CheckPermissionRequest):**


``` protobuf
message CheckPermissionRequest {
  string user_id = 1;              // ID пользователя
  string resource = 2;              // ресурс (например "device:dev-123")
  string action = 3;                 // действие ("update", "delete", "control")
  map<string, string> context = 4;   // доп. контекст (опционально)
}

```
**Пример запроса**

```json
{
  "user_id": "user-123",
  "resource": "device:dev-123",
  "action": "control",
  "context": {
    "device_type": "light",
    "room": "living_room"
  }
}
```

**Response (CheckPermissionResponse):**

``` protobuf
message CheckPermissionResponse {
  bool allowed = 1;                  // разрешено/запрещено
  string reason = 2;                  // причина запрета (если allowed = false)
  repeated string grants = 3;         // какие права сработали
  int32 ttl_seconds = 4;               // время жизни проверки (для кэширования)
}

```
**Пример ответа**

```json
{
  "allowed": true,
  "grants": ["owner", "room_access"],
  "ttl_seconds": 300
}

```

**Описание полей**

|    Поле       |   Тип   |            Описание            |
|---------------|---------|--------------------------------|
| `allowed`     | boolean | `true` если действие разрешено |
| `reason`      | string  | Причина отказа (для ошибок)    |
| `grants`      | array   | Какие права сработали          |
| `ttl_seconds` | int     | Время кэширования результата   |


### 5.2 Device Service → Audit Log (запись действия)

**Сервис:** Kafka топик `audit-events`
**Протокол:** Асинхронное событие

**Сообщение:**

```json
{
  "eventId": "evt-456",
  "eventType": "device.command",
  "timestamp": "2026-02-20T14:30:00Z",
  "userId": "user-123",
  "deviceId": "dev-123",
  "action": "turn_on",
  "result": "success",
  "metadata": {
    "ip": "192.168.1.100",
    "userAgent": "iOS App/2.3.4"
  }
}
```

## 6. Логика проверки прав (таблица)

| Роль пользователя | Действие с устройством |       Результат       |
|-------------------|------------------------|-----------------------|
| Владелец          | Любое действие         |  Разрешено            |
| Член семьи        | Включить/выключить     |  Разрешено            |
| Член семьи        | Удалить устройство     |  Запрещено            |
| Гость             | Включить свет          |  Разрешено (временно) |
| Гость             | Изменить настройки     |  Запрещено            |
| Посторонний       | Любое действие         |  Запрещено            |

## 7. Обработка ошибок

| HTTP код |        Описание            |                  Пример ответа                                      |
|----------|----------------------------|---------------------------------------------------------------------|
|   400    | Неверный формат запроса    | `{"error": "Invalid command format"}`                               |
|   401    | Не авторизован             | `{"error": "Token expired"}`                                        |
|   403    | Нет прав                   | `{"error": "User does not have permission to control this device"}` |
|   404    | Устройство не найдено      | `{"error": "Device not found"}`                                     |
|   409    | Конфликт версий            | `{"error": "Device state changed, please retry"}`                   |
|   412    | Условие не выполнено       | `{"error": "If-Match condition failed"}`                            |
|   429    | Слишком много запросов     | `{"error": "Rate limit exceeded. Try again in 30 seconds"}`         |
|   503    | Сервис временно недоступен | `{"error": "Auth service unavailable, please retry later"}`         |

## 8. Кэширование прав

Для оптимизации Device Service может кэшировать результаты проверки прав:

```python
# Псевдокод логики кэширования
cache_key = f"permissions:{user_id}:{resource}:{action}"
cached = cache.get(cache_key)

if cached and cached.ttl > 0:
    return cached.allowed
else:
    result = auth_service.CheckPermission(request)
    cache.set(cache_key, result, ttl=result.ttl_seconds)
    return result.allowed

```

## 9. Диаграмма последовательноси (текстовая)

┌─────────┐          ┌───────────────┐          ┌─────────────┐          ┌────────────┐
│ Клиент  │          │ Device Service|          │ Auth Service│          │ Audit Log  │
└────┬────┘          └───────┬───────┘          └──────┬──────┘          └─────┬──────┘
     │                       │                         │                       │
     │ PATCH /devices/123    │                         │                       │
     │──────────────────────>│                         │                       │
     │                       │                         │                       │
     │                       │ CheckPermission()       │                       │
     │                       │────────────────────────>│                       │
     │                       │                         │                       │
     │                       │ PermissionResponse      │                       │
     │                       │<────────────────────────│                       │
     │                       │                         │                       │
     │                       │ если разрешено:         │                       │
     │                       │ ──► выполнить команду   │                       │
     │                       │                         │                       │
     │                       │ publish audit event     │                       │
     │                       │────────────────────────────────────────────────>│
     │                       │                         │                       │
     │ Response 200 OK       │                         │                       │
     │<──────────────────────│                         │                       │
     │                       │                         │                       │

## 10. Примечания по безопасности
1. Токен должен быть валидным — проверяется в Auth Service
2. Права проверяются на каждый запрос — нельзя полагаться на кэш для критичных действий
3. Аудит всех действий — обязательно для безопасности
4. Rate limiting — защита от брутфорса
5. Idempotency — повторные запросы не должны менять состояние дважды

## 11. Визуализация
[Диаграмма последовательности проверки прав](../diagrams/02-check-permission-sequence.png)