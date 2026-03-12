# Сценарий 5: Отправка уведомления пользователю

## 1. Участники

- **Клиент**: мобильное приложение пользователя / веб-интерфейс
- **Notification Service**: сервис уведомлений
- **Auth Service**: сервис аутентификации
- **Device Service**: сервис управления устройствами
- **Automation Service**: сервис автоматизации
- **Kafka**: шина событий (для асинхронных уведомлений)
- **Firebase/APNs**: внешние push-сервисы
- **SMTP Server**: почтовый сервер

## 2. Описание

Пользователь получает уведомления о событиях в системе умного дома: срабатывание датчиков, выполнение сценариев, низкий заряд батареи, отключение устройств и т.д. Уведомления могут отправляться через разные каналы: push-уведомления, email, SMS или встроенные уведомления в приложении.

## 3. Последовательность шагов

1. Сервис-источник (Device, Automation, Billing и др.) публикует событие в Kafka
2. Notification Service подписан на соответствующие топики и получает событие
3. Notification Service проверяет настройки уведомлений пользователя (включён ли канал, не в тихом часу)
4. Notification Service выбирает шаблон уведомления в зависимости от типа события
5. Notification Service рендерит шаблон с данными из события
6. Notification Service отправляет уведомление через соответствующий канал:
   - Push: через Firebase (Android) или APNs (iOS)
   - Email: через SMTP-сервер
   - In-App: через WebSocket в приложении
7. Notification Service сохраняет историю уведомления в базу данных
8. Клиент получает уведомление и отображает пользователю

## 4. Детали запроса от клиента

### 4.1 Получение истории уведомлений

**Endpoint:** `GET /api/v1/notifications`

### Заголовки (Headers)

|     Header    |      Значение      | Обязательный |        Описание         |
|---------------|--------------------|--------------|-------------------------|
| Authorization | `Bearer {token}`   |     Да       | JWT токен пользователя  |
| Accept        | `application/json` |     Нет      | Ожидаемый формат ответа |

### Параметры запроса (Query Parameters)

|    Параметр   |   Тип   | Обязательный |       Описание         |         Пример            |
|---------------|---------|--------------|------------------------|---------------------------|
| `limit`       | integer |      Нет     | Количество уведомлений | `50`                      |
| `offset`      | integer |      Нет     | Смещение для пагинации | `0`                       |
| `unread_only` | boolean |      Нет     | Только непрочитанные   | `true`                    |
| `channel`     | string  |      Нет     | Фильтр по каналу       | `push`, `email`, `in_app` |
| `from`        | string  |      Нет     | Начальная дата         | `2026-02-01T00:00:00Z`    |
| `to`          | string  |      Нет     | Конечная дата          | `2026-02-22T23:59:59Z`    |

### Пример полного запроса

```http
GET /api/v1/notifications?limit=20&unread_only=true HTTP/1.1
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
  "notifications": [
    {
      "id": "notif-123",
      "user_id": "user-123",
      "channel": "push",
      "type": "motion_detected",
      "title": "Движение в гостиной",
      "body": "Датчик движения обнаружил активность в 22:30",
      "data": {
        "device_id": "sensor-living",
        "room": "living_room",
        "timestamp": "2026-02-21T22:30:00Z"
      },
      "priority": "high",
      "status": "delivered",
      "read": false,
      "created_at": "2026-02-21T22:30:05Z",
      "delivered_at": "2026-02-21T22:30:06Z"
    },
    {
      "id": "notif-122",
      "user_id": "user-123",
      "channel": "email",
      "type": "weekly_report",
      "title": "Еженедельный отчёт",
      "body": "Ваш еженедельный отчёт по能耗 готов",
      "data": {
        "report_url": "https://api.smarthome.com/reports/week-08.pdf"
      },
      "priority": "normal",
      "status": "sent",
      "read": true,
      "created_at": "2026-02-21T08:00:00Z",
      "read_at": "2026-02-21T09:15:00Z"
    }
  ],
  "total": 2,
  "unread_count": 1,
  "limit": 20,
  "offset": 0
}
```
**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1notifications/get)
- **gRPC:** [`auth.AuthService/GetUserPreferences`](../proto/auth.proto)


### 4.2 Отметка уведомлений как прочитанных

**Endpoint:** `POST /api/v1/notifications/mark-read`

### Тело запроса

```json
{
  "notification_ids": ["notif-123", "notif-124"],
  "mark_all": false
}
```

### Ответ (200 OK)

```json
{
  "marked_count": 2,
  "unread_count": 0
}
```
**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1notifications~1mark-read/post)

### 4.3 Настройка предпочтений уведомлений

**Endpoint:** `PUT /api/v1/notifications/preferences`

### Тело запроса

```json
{
  "channels": {
    "push": {
      "enabled": true,
      "quiet_hours_start": "22:00",
      "quiet_hours_end": "08:00"
    },
    "email": {
      "enabled": true,
      "address": "user@example.com"
    },
    "sms": {
      "enabled": false,
      "phone": "+1234567890"
    },
    "in_app": {
      "enabled": true
    }
  },
  "categories": {
    "security_alerts": {
      "push": true,
      "email": true,
      "sms": true,
      "in_app": true,
      "min_priority": "high"
    },
    "device_events": {
      "push": true,
      "email": false,
      "in_app": true,
      "min_priority": "normal"
    },
    "weekly_reports": {
      "email": true,
      "in_app": true,
      "min_priority": "low"
    },
    "marketing": {
      "email": false,
      "in_app": false
    }
  }
}
```

### Ответ (200 OK)

```json
{
  "user_id": "user-123",
  "updated_at": "2026-02-22T10:00:00Z",
  "preferences": { ... }
}
```

**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1notifications~1preferences/put)
- **gRPC:** [`auth.AuthService/GetUserPreferences`](../proto/auth.proto)

## 5. Межсервисное взаимодействие

### 5.1 Kafka → Notification Service (получение событий)

**Топики:** `device.events`, `automation.execution`, `billing.events`, `security.alerts`

**Пример сообщения из топика `device.events`:**

```json
{
  "eventId": "evt-dev-789",
  "eventType": "device.offline",
  "timestamp": "2026-02-22T14:30:00Z",
  "deviceId": "temp-123",
  "deviceName": "Living Room Sensor",
  "userId": "user-123",
  "data": {
    "last_seen": "2026-02-22T14:25:00Z",
    "reason": "timeout",
    "battery": 15
  }
}
```

**Пример сообщения из топика `automation.execution`:**

```json
{
  "eventId": "evt-auto-456",
  "eventType": "scenario.completed",
  "timestamp": "2026-02-22T08:00:00Z",
  "scenarioId": "scenario-123",
  "scenarioName": "Morning Light",
  "userId": "user-123",
  "data": {
    "executionId": "exec-789",
    "duration_ms": 3200,
    "actions_count": 2,
    "success": true
  }
}
```

### 5.2 Notification Service → User Preferences (получение настроек)

**Сервис:** `auth.AuthService`
**Метод:** `GetUserPreferences`
**Протокол:** gRPC

**Request (GetUserPreferencesRequest):**
```protobuf
message GetUserPreferencesRequest {
  string user_id = 1;
  repeated string preference_categories = 2;  // опционально
}
```

**Response (GetUserPreferencesResponse):**
```protobuf
message GetUserPreferencesResponse {
  string user_id = 1;
  map<string, ChannelPreferences> channels = 2;
  map<string, CategoryPreferences> categories = 3;
  google.protobuf.Timestamp updated_at = 4;
}

message ChannelPreferences {
  bool enabled = 1;
  string quiet_hours_start = 2;
  string quiet_hours_end = 3;
  string address = 4;  // email, phone и т.д.
}

message CategoryPreferences {
  bool push = 1;
  bool email = 2;
  bool sms = 3;
  bool in_app = 4;
  string min_priority = 5;  // "low", "normal", "high", "urgent"
}
```

### 5.3 Notification Service → Template Service (рендеринг шаблонов)

**Сервис:** Внутренний компонент Notification Service
**Метод:** `RenderTemplate`

**Шаблоны хранятся в базе данных:**

| ID |     Name       | Channel |            Subject              |                               Body                                   |            Variables             |
|----|----------------|---------|---------------------------------|----------------------------------------------------------------------|----------------------------------|
| 1  | motion_alert   |   push  | "Движение в {{room}}"           | "Датчик движения обнаружил активность в {{time}}"                    | ["room", "time"]                 |
| 2  | device_offline |   push  | "Устройство отключилось"        | "{{deviceName}} не отвечает. Последняя активность: {{lastSeen}}"     | ["deviceName", "lastSeen"]       |
| 3  | weekly_report  |   email | "Еженедельный отчёт - {{week}}" | "Ваш отчёт за неделю {{week}} готов. {{summary}}"                    | ["week", "summary", "reportUrl"] |
| 4  | battery_low    |   push  | "Низкий заряд батареи"          | "Устройство {{deviceName}} имеет {{battery}}% заряда"                | ["deviceName", "battery"]        |

**Процесс рендеринга:**
```python
def render_notification(template_name, channel, data, locale="ru"):
    # Получить шаблон из БД
    template = db.get_template(template_name, channel, locale)
    
    # Рендеринг с использованием Jinja2/Handlebars
    subject = render_template(template.subject, data)
    body = render_template(template.body, data)
    
    return {
        "subject": subject,
        "body": body,
        "template_id": template.id
    }
```

### 5.4 Notification Service → Firebase/APNs (отправка push)

**Сервис:** Firebase Cloud Messaging (Android) / Apple Push Notification Service (iOS)

**Запрос к FCM (пример):**

```http
POST https://fcm.googleapis.com/v1/projects/smarthome/messages:send
Authorization: Bearer ya29.c.Elq...
Content-Type: application/json
```

```json
{
  "message": {
    "token": "fcm_device_token_123",
    "notification": {
      "title": "Движение в гостиной",
      "body": "Датчик движения обнаружил активность в 22:30"
    },
    "data": {
      "type": "motion_alert",
      "device_id": "sensor-living",
      "room": "living_room",
      "timestamp": "2026-02-21T22:30:00Z",
      "click_action": "OPEN_CAMERA"
    },
    "android": {
      "priority": "high",
      "notification": {
        "channel_id": "alerts",
        "sound": "default",
        "vibrate_timings": ["0s", "1s"]
      }
    },
    "apns": {
      "payload": {
        "aps": {
          "alert": {
            "title": "Движение в гостиной",
            "body": "Датчик движения обнаружил активность в 22:30"
          },
          "sound": "default",
          "badge": 1,
          "category": "ALERT"
        }
      }
    }
  }
}
```

### 5.5 Notification Service → SMTP Server (отправка email)

**Сервис:** SMTP-сервер (SendGrid, Mailgun, или собственный)

**SMTP-команды:**
```
HELO smtp.smarthome.com
MAIL FROM: notifications@smarthome.com
RCPT TO: user@example.com
DATA
From: Smart Home <notifications@smarthome.com>
To: user@example.com
Subject: Еженедельный отчёт - Неделя 8
Content-Type: text/html; charset=utf-8

<html>
<body>
<h2>Ваш еженедельный отчёт</h2>
<p>За неделю 8 (15-21 февраля 2026):</p>
<ul>
  <li>Средняя температура: 23.5°C</li>
  <li>Потребление энергии: 45.2 кВт·ч</li>
  <li>Срабатываний датчиков: 128</li>
</ul>
<p><a href="https://api.smarthome.com/reports/week-08.pdf">Скачать полный отчёт</a></p>
</body>
</html>
.
QUIT
```

### 5.6 Notification Service → WebSocket (In-App уведомления)

**Сервис:** WebSocket-сервер (для онлайн-клиентов)

**Сообщение через WebSocket:**

```json
{
  "type": "notification",
  "data": {
    "id": "notif-123",
    "title": "Движение в гостиной",
    "body": "Датчик движения обнаружил активность в 22:30",
    "priority": "high",
    "timestamp": "2026-02-21T22:30:05Z",
    "data": {
      "device_id": "sensor-living",
      "room": "living_room"
    }
  }
}
```

**Полные спецификации:**
- **gRPC:** [`auth.proto`](../proto/auth.proto)
- **Kafka:** [`device.events`](../asyncapi/asyncapi.yaml#/channels/device.events)
- **Kafka:** [`automation.execution`](../asyncapi/asyncapi.yaml#/channels/automation.execution)
- **Kafka:** [`billing.events`](../asyncapi/asyncapi.yaml#/channels/billing.events)
- **Kafka:** [`security.alerts`](../asyncapi/asyncapi.yaml#/channels/security.alerts)

## 6. Типы уведомлений и шаблоны

### 6.1 Push-уведомления

|         Тип          | Приоритет |           Заголовок          |                                Тело                                |      Действие      |
|----------------------|-----------|------------------------------|--------------------------------------------------------------------|--------------------|
| `motion_detected`    | high      | Движение в {{room}}          | Датчик движения обнаружил активность в {{time}}                    | Открыть камеру     |
| `device_offline`     | high      | {{deviceName}} отключилось   | Устройство не отвечает. Последняя активность: {{lastSeen}}         | Открыть устройство |
| `battery_low`        | normal    | Низкий заряд батареи         | Устройство {{deviceName}} имеет {{battery}}% заряда                | Открыть устройство |
| `door_open`          | high      | Дверь открыта                | {{doorName}} открыта уже {{minutes}} минут                         | Показать статус    |
| `scenario_completed` | normal    | Сценарий выполнен            | Сценарий "{{scenarioName}}" успешно выполнен                       | Открыть сценарий   |
| `payment_success`    | normal    | Платёж прошёл                | Оплата подписки на сумму {{amount}} {{currency}} успешно проведена | Открыть счёт       |
| `trial_ending`       | normal    | Заканчивается пробный период | Ваш пробный период заканчивается через {{days}} дней               | Продлить подписку  |

### 6.2 Email-уведомления

|        Тип         |                          Тема                            |               Содержание                 |
|--------------------|----------------------------------------------------------|------------------------------------------|
| `welcome_email`    | Добро пожаловать в Smart Home!                           | Приветственное письмо с инструкциями     |
| `weekly_report`    | Еженедельный отчёт - {{week}}                            | Статистика за неделю, графики, PDF-отчёт |
| `security_alert`   | Важно: Обнаружена подозрительная активность              | Детали события, рекомендации             |
| `invoice`          | Счёт №{{invoiceNumber}} на сумму {{amount}} {{currency}} | PDF-счёт, ссылка на оплату               |
| `password_changed` | Пароль изменён                                           | Уведомление о смене пароля               |

### 6.3 In-App уведомления

|      Тип  |    Отображается как | Длительность |
|-----------|---------------------|--------------|
| `info`    | Синее уведомление   | 5 секунд     |
| `success` | Зелёное уведомление | 5 секунд     |
| `warning` | Жёлтое уведомление  | 10 секунд    |
| `error`   | Красное уведомление | До закрытия  |

## 7. Обработка ошибок

| HTTP код |         Описание        |                   Пример ответа                             |
|----------|-------------------------|-------------------------------------------------------------|
|   400    | Неверный формат запроса | `{"error": "Invalid time format"}`                          |
|   401    | Не авторизован          | `{"error": "Token expired"}`                                |
|   403    | Нет прав                | `{"error": "Cannot modify another user's preferences"}`     |
|   404    | Уведомление не найдено  | `{"error": "Notification not found"}`                       |
|   422    | Неверные параметры      | `{"error": "Invalid quiet hours format"}`                   |
|   429    | Слишком много запросов  | `{"error": "Rate limit exceeded. Try again in 60 seconds"}` |
|   500    | Внутренняя ошибка       | `{"error": "Failed to render template"}`                    |
|   502    | Ошибка внешнего сервиса | `{"error": "FCM service unavailable"}`                      |
|   503    | Сервис недоступен       | `{"error": "Email service temporarily unavailable"}`        |

## 8. Диаграмма последовательности (Mermaid)

```mermaid
sequenceDiagram
    participant Source as Сервис-источник
    participant Kafka as Kafka
    participant Notif as Notification Service
    participant Auth as Auth Service
    participant DB as PostgreSQL
    participant FCM as Firebase/APNs
    participant SMTP as SMTP Server
    participant Client as Клиент (App)

    Source->>Kafka: publish event
    activate Kafka
    Kafka->>Notif: consume event
    deactivate Kafka
    
    activate Notif
    
    Notif->>Auth: GetUserPreferences(user_id)
    activate Auth
    Auth-->>Notif: {channels: {...}, categories: {...}}
    deactivate Auth
    
    Notif->>Notif: Проверка quiet hours
    Notif->>Notif: Выбор шаблона
    Notif->>Notif: Рендеринг шаблона
    
    alt Push-уведомление
        Notif->>FCM: send push notification
        activate FCM
        FCM-->>Client: push received
        FCM-->>Notif: delivery receipt
        deactivate FCM
    else Email-уведомление
        Notif->>SMTP: send email
        activate SMTP
        SMTP-->>Notif: email accepted
        deactivate SMTP
    else In-App уведомление (если онлайн)
        Notif->>Client: WebSocket message
        activate Client
        Client-->>Notif: delivered
        deactivate Client
    end
    
    Notif->>DB: save notification history
    activate DB
    DB-->>Notif: saved
    deactivate DB
    
    deactivate Notif
```

## 9. Очереди и повторные попытки

### 9.1 Очередь уведомлений

Notification Service использует очередь для обработки уведомлений:

```python
# Псевдокод обработки очереди
class NotificationQueue:
    def __init__(self):
        self.queue = []
        self.max_retries = 3
        
    def add(self, notification):
        notification.retry_count = 0
        self.queue.append(notification)
        
    def process(self):
        while self.queue:
            notification = self.queue.pop(0)
            success = self.send(notification)
            
            if not success and notification.retry_count < self.max_retries:
                notification.retry_count += 1
                # Повторная попытка через 5, 30, 120 секунд
                delay = [5, 30, 120][notification.retry_count - 1]
                schedule_retry(notification, delay)
```

### 9.2 Стратегии повторных попыток

| Канал | Макс. попыток |         Интервалы (сек)     |        Причина отказа         |
|-------|---------------|-----------------------------|-------------------------------|
| Push  |       3       | 5, 30, 120                  | Устройство офлайн, ошибка FCM |
| Email |       5       | 10, 60, 300, 3600, 86400    | SMTP временно недоступен      |
| SMS   |       2       | 60, 300                      | Ошибка провайдера            |

## 10. Логирование и мониторинг

### 10.1 Логи отправки

```json
{
  "timestamp": "2026-02-22T14:30:05Z",
  "level": "INFO",
  "service": "notification-service",
  "notification_id": "notif-123",
  "user_id": "user-123",
  "channel": "push",
  "type": "motion_detected",
  "status": "sent",
  "provider": "FCM",
  "provider_response": "message_id: 0:123456789",
  "duration_ms": 150
}
```

### 10.2 Метрики

|         Метрика            |      Описание      | Где смотреть |
|----------------------------|--------------------|--------------|
| `notifications.sent.total` | Всего отправлено   | Prometheus   |
| `notifications.sent.push`  | Push-уведомлений   | Prometheus   |
| `notifications.sent.email` | Email-уведомлений  | Prometheus   |
| `notifications.sent.sms`   | SMS-уведомлений    | Prometheus   |
| `notifications.delivered`  | Доставлено успешно | Prometheus   |
| `notifications.failed`     | Ошибок доставки    | Prometheus   |
| `notifications.retries`    | Повторных попыток  | Prometheus   |
| `notifications.latency_ms` | Время доставки     | Prometheus   |
| `notifications.queue.size` | Размер очереди     | Prometheus   |

## 11. Безопасность и конфиденциальность

1. **Персональные данные** — email, телефон хранятся в зашифрованном виде
2. **Токены устройств** — обновляются при каждой установке приложения
3. **Отписка** — пользователь может отписаться от любых каналов
4. **Тихие часы** — уведомления не отправляются в указанный период
5. **Лимиты** — не более 100 push/день, 10 email/день на пользователя
6. **GDPR compliance** — возможность экспорта/удаления истории уведомлений

## 12. Примеры использования

### Пример 1: Уведомление о движении

**Событие от Device Service:**
```json
{
  "eventType": "motion.detected",
  "deviceId": "sensor-living",
  "userId": "user-123",
  "timestamp": "2026-02-22T22:30:00Z",
  "data": {
    "room": "living_room",
    "confidence": 95
  }
}
```

**Сгенерированное push-уведомление:**
```json
{
  "title": "Движение в гостиной",
  "body": "Датчик движения обнаружил активность в 22:30",
  "data": {
    "type": "motion_alert",
    "device_id": "sensor-living",
    "room": "living_room",
    "timestamp": "2026-02-22T22:30:00Z"
  }
}
```

### Пример 2: Еженедельный отчёт по email

```html
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .stats { display: flex; justify-content: space-around; margin: 20px 0; }
        .stat-card { background: #f5f5f5; padding: 15px; border-radius: 5px; text-align: center; }
        .stat-value { font-size: 24px; font-weight: bold; color: #4CAF50; }
        .stat-label { color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Еженедельный отчёт - Неделя 8</h1>
        <p>15 - 21 февраля 2026</p>
    </div>
    
    <div class="stats">
        <div class="stat-card">
            <div class="stat-value">23.5°C</div>
            <div class="stat-label">Средняя температура</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">45.2 кВт·ч</div>
            <div class="stat-label">Потребление энергии</div>
        </div>
        <div class="stat-card">
            <div class="stat-value">128</div>
            <div class="stat-label">Событий</div>
        </div>
    </div>
    
    <h3>Подробная статистика по комнатам:</h3>
    <table style="width: 100%; border-collapse: collapse;">
        <tr style="background-color: #f2f2f2;">
            <th style="padding: 10px; text-align: left;">Комната</th>
            <th style="padding: 10px; text-align: left;">Температура</th>
            <th style="padding: 10px; text-align: left;">Влажность</th>
            <th style="padding: 10px; text-align: left;">Активность</th>
        </tr>
        <tr>
            <td style="padding: 10px;">Гостиная</td>
            <td style="padding: 10px;">23.5°C</td>
            <td style="padding: 10px;">45%</td>
            <td style="padding: 10px;">56 событий</td>
        </tr>
        <tr>
            <td style="padding: 10px;">Спальня</td>
            <td style="padding: 10px;">22.1°C</td>
            <td style="padding: 10px;">48%</td>
            <td style="padding: 10px;">32 события</td>
        </tr>
    </table>
    
    <p style="margin-top: 30px;">
        <a href="https://api.smarthome.com/reports/week-08.pdf" 
           style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
            Скачать полный отчёт (PDF)
        </a>
    </p>
    
    <p style="color: #666; font-size: 12px; margin-top: 30px;">
        Вы получили это письмо, потому что подписаны на еженедельные отчёты.
        <a href="https://api.smarthome.com/unsubscribe?token=...">Отписаться</a>
    </p>
</body>
</html>
```

### Пример 3: Уведомление о низком заряде батареи

**Событие от Device Service:**
```json
{
  "eventType": "battery.low",
  "deviceId": "temp-bedroom",
  "userId": "user-123",
  "timestamp": "2026-02-22T18:00:00Z",
  "data": {
    "deviceName": "Bedroom Sensor",
    "battery": 15,
    "estimated_days_left": 7
  }
}
```

**Сгенерированное уведомление:**
```json
{
  "title": "Низкий заряд батареи",
  "body": "Устройство Bedroom Sensor имеет 15% заряда. Замените батарею в ближайшее время.",
  "priority": "normal",
  "data": {
    "type": "battery_alert",
    "device_id": "temp-bedroom",
    "battery": 15
  }
}
```

## 13. Связанные спецификации

- **REST API:** [`openapi.yaml`](../openapi/openapi.yaml)
- **gRPC API:** [`auth.proto`](../proto/auth.proto)
- **AsyncAPI:** [`asyncapi.yaml`](../asyncapi/asyncapi.yaml)

