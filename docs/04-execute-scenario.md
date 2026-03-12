# Сценарий 4: Выполнение сценария автоматизации

## 1. Участники

- **Клиент**: мобильное приложение пользователя / веб-интерфейс
- **Automation Service**: сервис управления сценариями
- **Device Service**: сервис управления устройствами
- **Integration Service**: сервис интеграции с устройствами
- **Notification Service**: сервис уведомлений
- **Auth Service**: сервис аутентификации
- **Kafka**: шина событий (для асинхронных уведомлений)

## 2. Описание

Пользователь создаёт сценарий автоматизации (например, "включить свет в 8 утра" или "если температура > 25°, включить кондиционер"). Сценарий может запускаться автоматически по триггеру или вручную из приложения. При срабатывании Automation Service проверяет условия, выполняет действия и уведомляет пользователя о результате.

## 3. Последовательность шагов (ручной запуск)

1. Пользователь нажимает "Выполнить сценарий" в приложении
2. Приложение отправляет POST-запрос в Automation Service
3. Automation Service проверяет права пользователя через Auth Service
4. Automation Service получает список устройств из сценария
5. Для каждого устройства Automation Service проверяет его доступность через Device Service
6. Automation Service отправляет команды устройствам через Integration Service
7. Integration Service передаёт команды реальным устройствам (MQTT/HTTP/Zigbee)
8. После выполнения всех действий Automation Service публикует событие в Kafka
9. Notification Service получает событие и отправляет уведомление пользователю
10. Automation Service возвращает ответ клиенту с результатами выполнения

## 4. Детали запроса от клиента

### 4.1 Создание сценария

**Endpoint:** `POST /api/v1/scenarios`

### Заголовки (Headers)

|      Header   |       Значение     | Обязательный |         Описание       |
|---------------|--------------------|--------------|------------------------|
| Authorization | `Bearer {token}`   |      Да      | JWT токен пользователя |
| Content-Type  | `application/json` |      Да      | Формат данных          |

### Тело запроса (Request Body)

```json
{
  "name": "Morning Light",
  "description": "Включить свет в спальне в 8 утра",
  "enabled": true,
  "triggers": [
    {
      "type": "time",
      "cron": "0 8 * * *",
      "timezone": "Europe/Moscow"
    }
  ],
  "conditions": [
    {
      "type": "device_state",
      "device_id": "sensor-bedroom",
      "property": "presence",
      "operator": "eq",
      "value": true
    }
  ],
  "actions": [
    {
      "type": "device_command",
      "device_id": "light-bedroom",
      "command": "turn_on",
      "parameters": {
        "brightness": 80,
        "transition_ms": 3000
      }
    },
    {
      "type": "notification",
      "template": "scenario_executed",
      "data": {
        "message": "Доброе утро! Свет включён."
      }
    }
  ]
}

```

### Описание полей запроса

|     Поле     |   Тип   | Обязательный |              Описание              |
|--------------|---------|--------------|------------------------------------|
| `name`       | string  |     Да       | Название сценария                  |
| `description`| string  |     Нет      | Описание сценария                  |
| `enabled`    | boolean |     Да       | Активен ли сценарий                |
| `triggers`   | array   |     Да       | Список триггеров (условий запуска) |
| `conditions` | array   |     Нет      | Дополнительные условия             |
| `actions`    | array   |     Да       | Действия при срабатывании          |

### Типы триггеров

|      Тип       |                Параметры                       |            Описание            |
|----------------|------------------------------------------------|--------------------------------|
| `time`         | `cron`, `timezone`                             | Расписание (cron-выражение)    |
| `device_state` | `device_id`, `property`, `operator`, `value`   | Изменение состояния устройства |
| `telemetry`    | `device_id`, `metric`, `operator`, `threshold` | Достижение порога телеметрии   |
| `sun`          | `event` (`sunrise`/`sunset`), `offset_minutes` | Восход/закат солнца            |
| `manual`       | -                                              | Только ручной запуск           |

### Типы действий

|        Тип       |               Параметры              |        Описание         |
|------------------|--------------------------------------|-------------------------|
| `device_command` | `device_id`, `command`, `parameters` | Команда устройству      |
| `delay`          | `seconds`                            | Пауза между действиями  |
| `notification`   | `template`, `data`                   | Отправка уведомления    |
| `webhook`        | `url`, `method`, `body`              | Вызов внешнего API      |
| `scene`          | `scene_id`                           | Запуск другого сценария |

### Ответ при успехе (201 Created)

```http
HTTP/1.1 201 Created
Content-Type: application/json
Location: /api/v1/scenarios/scenario-123
```

```json
{
  "id": "scenario-123",
  "name": "Morning Light",
  "enabled": true,
  "triggers": [
    {
      "type": "time",
      "cron": "0 8 * * *",
      "timezone": "Europe/Moscow"
    }
  ],
  "actions": [
    {
      "type": "device_command",
      "device_id": "light-bedroom",
      "command": "turn_on"
    }
  ],
  "createdAt": "2026-02-21T10:00:00Z",
  "updatedAt": "2026-02-21T10:00:00Z"
}
```

**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1scenarios/post)
- **gRPC:** [`auth.AuthService/CheckScenarioPermission`](../proto/auth.proto)
- **gRPC:** [`device.DeviceService/GetDevicesBatch`](../proto/device.proto)
- **gRPC:** [`integration.IntegrationService/BatchSendCommands`](../proto/integration.proto)

### 4.2 Ручной запуск сценария

**Endpoint:** `POST /api/v1/scenarios/{scenarioId}/execute`

### Параметры пути (Path Parameters)

|    Параметр  |   Тип  | Обязательный |  Описание   |
|--------------|--------|--------------|-------------|
| `scenarioId` | string |      Да      | ID сценария |

### Пример запроса

```http
POST /api/v1/scenarios/scenario-123/execute HTTP/1.1
Host: api.smarthome.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyLTEyMyJ9...
Content-Type: application/json
```

```json
{
  "dry_run": false,
  "context": {
    "source": "manual",
    "reason": "user_requested"
  }
}
```

### Ответ при успехе (200 OK)

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "executionId": "exec-456",
  "scenarioId": "scenario-123",
  "scenarioName": "Morning Light",
  "status": "completed",
  "startedAt": "2026-02-21T08:00:00Z",
  "completedAt": "2026-02-21T08:00:03Z",
  "duration_ms": 3200,
  "trigger": "manual",
  "actions": [
    {
      "type": "device_command",
      "device_id": "light-bedroom",
      "command": "turn_on",
      "status": "success",
      "message": "Command sent to device",
      "duration_ms": 150
    },
    {
      "type": "notification",
      "status": "success",
      "message": "Notification sent",
      "duration_ms": 50
    }
  ],
  "results": {
    "summary": "Все действия выполнены успешно",
    "success_count": 2,
    "failure_count": 0
  }
}
```
**Ссылки на спецификации:**
- **REST:** [`openapi.yaml`](../openapi/openapi.yaml#/paths/~1scenarios~1{scenarioId}~1execute/post)
- **gRPC:** [`auth.AuthService/CheckScenarioPermission`](../proto/auth.proto)
- **Kafka:** [`automation.execution`](../asyncapi/asyncapi.yaml#/channels/automation.execution)
- **Kafka:** [`notifications`](../asyncapi/asyncapi.yaml#/channels/notifications)

## 5. Межсервисное взаимодействие

### 5.1 Automation Service → Auth Service (проверка прав)

**Сервис:** `auth.AuthService`
**Метод:** `CheckScenarioPermission`
**Протокол:** gRPC

**Request (CheckScenarioPermissionRequest):**
```protobuf
message CheckScenarioPermissionRequest {
  string user_id = 1;
  string scenario_id = 2;
  string action = 3;  // "execute", "edit", "delete"
}
```

**Пример запроса:**
```json
{
  "user_id": "user-123",
  "scenario_id": "scenario-123",
  "action": "execute"
}
```

**Response:**
```json
{
  "allowed": true,
  "reason": "User is owner",
  "permissions": ["execute", "view"]
}
```

### 5.2 Automation Service → Device Service (проверка устройств)

**Сервис:** `device.DeviceService`
**Метод:** `GetDevicesBatch`
**Протокол:** gRPC

**Request (GetDevicesBatchRequest):**
```protobuf
message GetDevicesBatchRequest {
  repeated string device_ids = 1;
  bool check_online = 2;
}
```

**Пример запроса:**
```json
{
  "device_ids": ["light-bedroom", "sensor-bedroom"],
  "check_online": true
}
```

**Response:**
```json
{
  "devices": [
    {
      "device_id": "light-bedroom",
      "name": "Bedroom Light",
      "type": "light",
      "status": "online",
      "online": true
    },
    {
      "device_id": "sensor-bedroom",
      "name": "Bedroom Sensor",
      "type": "sensor",
      "status": "online",
      "online": true
    }
  ]
}
```

### 5.3 Automation Service → Integration Service (отправка команд)

**Сервис:** `integration.IntegrationService`
**Метод:** `BatchSendCommands`
**Протокол:** gRPC (синхронный для немедленного выполнения)

**Request (BatchSendCommandsRequest):**
```protobuf
message BatchSendCommandsRequest {
  repeated Command commands = 1;
  ExecutionMode mode = 2;  // SEQUENTIAL, PARALLEL
  int32 timeout_ms = 3;
}

message Command {
  string command_id = 1;
  string device_id = 2;
  string command = 3;
  google.protobuf.Struct parameters = 4;
  int32 priority = 5;  // 0-100
}
```

**Пример запроса:**
```json
{
  "commands": [
    {
      "command_id": "cmd-001",
      "device_id": "light-bedroom",
      "command": "turn_on",
      "parameters": {
        "brightness": 80
      },
      "priority": 50
    }
  ],
  "mode": "SEQUENTIAL",
  "timeout_ms": 5000
}
```

**Response:**
```json
{
  "results": [
    {
      "command_id": "cmd-001",
      "success": true,
      "status": "delivered",
      "response_time_ms": 120,
      "result": {
        "state": "on",
        "brightness": 80
      }
    }
  ],
  "overall_success": true,
  "total_duration_ms": 150
}
```

### 5.4 Automation Service → Kafka (публикация события выполнения)

**Топик:** `automation.execution`

**Сообщение:**
```json
{
  "eventId": "evt-exec-789",
  "eventType": "scenario.executed",
  "timestamp": "2026-02-21T08:00:03Z",
  "scenarioId": "scenario-123",
  "scenarioName": "Morning Light",
  "userId": "user-123",
  "trigger": "manual",
  "status": "success",
  "executionTime": 3200,
  "actions": [
    {
      "type": "device_command",
      "deviceId": "light-bedroom",
      "success": true
    }
  ],
  "metadata": {
    "source": "mobile_app",
    "app_version": "2.3.4"
  }
}
```

### 5.5 Automation Service → Notification Service (уведомление)

**Сервис:** `notification.NotificationService`
**Метод:** `SendNotification` (асинхронно, через Kafka)

**Сообщение в Kafka (топик `notifications`):**
```json
{
  "notificationId": "notif-567",
  "userId": "user-123",
  "channel": "push",
  "title": "Сценарий выполнен",
  "body": "Сценарий 'Morning Light' успешно выполнен",
  "data": {
    "scenarioId": "scenario-123",
    "executionId": "exec-456",
    "screen": "Scenarios"
  },
  "priority": "normal"
}
```
**Полные спецификации:**
- **gRPC:** [`auth.proto`](../proto/auth.proto)
- **gRPC:** [`device.proto`](../proto/device.proto)
- **gRPC:** [`integration.proto`](../proto/integration.proto)
- **Kafka:** [`asyncapi.yaml`](../asyncapi/asyncapi.yaml)

## 6. Движок правил (Rule Engine)

### 6.1 Оценка условий

```python
# Псевдокод логики оценки условий
def evaluate_condition(condition, context):
    if condition.type == "device_state":
        device = get_device_state(condition.device_id)
        value = device.get(condition.property)
        return compare(value, condition.operator, condition.value)
    
    elif condition.type == "time":
        now = get_current_time(condition.timezone)
        return matches_cron(now, condition.cron)
    
    elif condition.type == "and":
        return all(evaluate_condition(c, context) for c in condition.conditions)
    
    elif condition.type == "or":
        return any(evaluate_condition(c, context) for c in condition.conditions)
    
    elif condition.type == "not":
        return not evaluate_condition(condition.condition, context)
    
    return False
```

### 6.2 Очередь выполнения действий

```python
# Псевдокод выполнения действий
async def execute_actions(actions):
    results = []
    
    for action in actions:
        if action.type == "delay":
            await asyncio.sleep(action.seconds)
            results.append({"type": "delay", "status": "completed"})
        
        elif action.type == "device_command":
            result = await integration_service.send_command(
                action.device_id,
                action.command,
                action.parameters
            )
            results.append({
                "type": "device_command",
                "device_id": action.device_id,
                "status": "success" if result.success else "failed",
                "result": result
            })
        
        elif action.type == "parallel":
            # Параллельное выполнение
            tasks = [execute_actions(a) for a in action.actions]
            parallel_results = await asyncio.gather(*tasks, return_exceptions=True)
            results.extend(parallel_results)
    
    return results
```

## 7. Обработка ошибок

| HTTP код |         Описание         |                Пример ответа                  |
|----------|--------------------------|-----------------------------------------------|
|   400    | Неверный формат запроса  | `{"error": "Invalid cron expression"}`        |
|   401    | Не авторизован           | `{"error": "Token expired"}`                  |
|   403    | Нет прав на выполнение   | `{"error": "User does not own this scenario"}`|
|   404    | Сценарий не найден       | `{"error": "Scenario not found"}`             |
|   409    | Конфликт выполнения      | `{"error": "Scenario already running"}`       |
|   422    | Неверные параметры       | `{"error": "Device light-bedroom not found"}` |
|   424    | Зависимость не выполнена | `{"error": "Previous action failed"}`         |
|   429    | Слишком много запросов   | `{"error": "Too many executions, try later"}` |
|   500    | Внутренняя ошибка        | `{"error": "Rule engine error"}`              |
|   503    | Сервис недоступен        | `{"error": "Device service unavailable"}`     |

## 8. Диаграмма последовательности (Mermaid)

```mermaid
sequenceDiagram
    participant Client as Клиент (App)
    participant Auto as Automation Service
    participant Auth as Auth Service
    participant Device as Device Service
    participant Integ as Integration Service
    participant MQTT as MQTT Broker
    participant Kafka as Kafka
    participant Notif as Notification Service

    Client->>Auto: POST /scenarios/{id}/execute
    activate Auto
    
    Auto->>Auth: CheckScenarioPermission(user, scenario)
    activate Auth
    Auth-->>Auto: {allowed: true}
    deactivate Auth
    
    Auto->>Auto: Загрузка сценария из БД
    Note right of Auto: Сценарий содержит действия
    
    Auto->>Device: GetDevicesBatch(device_ids)
    activate Device
    Device-->>Auto: {devices: [...]}
    deactivate Device
    
    loop Для каждого действия
        Auto->>Integ: BatchSendCommands(commands)
        activate Integ
        
        Integ->>MQTT: publish command
        activate MQTT
        MQTT-->>Integ: command sent
        deactivate MQTT
        
        MQTT->>Integ: device response
        activate Integ
        Integ-->>Auto: {success: true, result: {...}}
        deactivate Integ
    end
    
    Auto->>Kafka: publish automation.execution
    activate Kafka
    Kafka-->>Auto: event published
    deactivate Kafka
    
    Kafka->>Notif: consume automation.execution
    activate Notif
    Notif-->>Client: push notification
    deactivate Notif
    
    Auto-->>Client: {executionId: "...", status: "completed"}
    deactivate Auto
```

## 9. Примеры сценариев

### Сценарий А: Умное освещение по времени

```json
{
  "name": "Вечерний режим",
  "triggers": [
    {
      "type": "sun",
      "event": "sunset",
      "offset_minutes": -30
    }
  ],
  "conditions": [
    {
      "type": "device_state",
      "device_id": "sensor-living",
      "property": "presence",
      "operator": "eq",
      "value": true
    }
  ],
  "actions": [
    {
      "type": "device_command",
      "device_id": "light-living",
      "command": "turn_on",
      "parameters": {
        "brightness": 70,
        "color_temp": 3000
      }
    },
    {
      "type": "device_command",
      "device_id": "light-dining",
      "command": "turn_on",
      "parameters": {
        "brightness": 60
      }
    },
    {
      "type": "delay",
      "seconds": 3600
    },
    {
      "type": "device_command",
      "device_id": "light-living",
      "command": "turn_off"
    }
  ]
}
```

### Сценарий Б: Климат-контроль

```json
{
  "name": "Автоматическая температура",
  "triggers": [
    {
      "type": "telemetry",
      "device_id": "sensor-living",
      "metric": "temperature",
      "operator": "gt",
      "threshold": 25,
      "for_seconds": 300
    }
  ],
  "actions": [
    {
      "type": "device_command",
      "device_id": "ac-living",
      "command": "turn_on",
      "parameters": {
        "mode": "cool",
        "temperature": 22,
        "fan_speed": "medium"
      }
    },
    {
      "type": "notification",
      "template": "ac_started",
      "data": {
        "room": "living",
        "temperature": 25
      }
    }
  ]
}
```

### Сценарий В: Безопасность

```json
{
  "name": "Оповещение о движении ночью",
  "triggers": [
    {
      "type": "device_state",
      "device_id": "sensor-motion",
      "property": "motion",
      "operator": "eq",
      "value": true
    }
  ],
  "conditions": [
    {
      "type": "time",
      "cron": "* 22-6 * * *",
      "timezone": "Europe/Moscow"
    },
    {
      "type": "device_state",
      "device_id": "alarm-system",
      "property": "state",
      "operator": "eq",
      "value": "armed"
    }
  ],
  "actions": [
    {
      "type": "parallel",
      "actions": [
        {
          "type": "device_command",
          "device_id": "light-outdoor",
          "command": "turn_on",
          "parameters": {
            "brightness": 100
          }
        },
        {
          "type": "notification",
          "template": "motion_alert",
          "data": {
            "location": "backyard",
            "time": "{{now}}"
          }
        },
        {
          "type": "webhook",
          "url": "https://api.security.com/alert",
          "method": "POST",
          "body": {
            "type": "motion",
            "location": "backyard"
          }
        }
      ]
    }
  ]
}
```

## 10. Метрики и мониторинг

Automation Service собирает метрики о работе:

|               Метрика              |        Описание        | Где смотреть |
|------------------------------------|------------------------|--------------|
| `automation.scenarios.total`       | Всего сценариев        |  Prometheus  |
| `automation.scenarios.active`      | Активных сценариев     |  Prometheus  |
| `automation.executions.total`      | Всего выполнений       |  Prometheus  |
| `automation.executions.success`    | Успешных выполнений    |  Prometheus  |
| `automation.executions.failed`     | Проваленных выполнений |  Prometheus  |
| `automation.execution.duration_ms` | Время выполнения       |  Prometheus  |
| `automation.trigger.evaluations`   | Оценок триггеров       |  Prometheus  |
| `automation.actions.sent`          | Отправленных действий  |  Prometheus  |

## 11. Логирование выполнения

```json
{
  "timestamp": "2026-02-21T08:00:03Z",
  "level": "INFO",
  "service": "automation-service",
  "execution_id": "exec-456",
  "scenario_id": "scenario-123",
  "user_id": "user-123",
  "trigger": "manual",
  "actions": [
    {
      "type": "device_command",
      "device_id": "light-bedroom",
      "success": true,
      "duration_ms": 150
    }
  ],
  "total_duration_ms": 3200,
  "status": "success"
}
```

## 12. Связанные спецификации

- **REST API:** [`openapi.yaml`](../openapi/openapi.yaml)
- **gRPC API:** [`auth.proto`](../proto/auth.proto), [`device.proto`](../proto/device.proto), [`integration.proto`](../proto/integration.proto)
- **AsyncAPI:** [`asyncapi.yaml`](../asyncapi/asyncapi.yaml)


