# Project_template

Это шаблон для решения проектной работы. Структура этого файла повторяет структуру заданий. Заполняйте его по мере работы над решением.

# Задание 1. Анализ и планирование



### 1. Описание функциональности монолитного приложения

**Управление отоплением:**

- Пользователи могут удалённо включать и выключать отопление в своих домах через веб-интерфейс.
- Система поддерживает прямое управление отоплением через отправку команд от сервера к датчикам.
- Каждое подключение требует выезда специалиста для настройки системы.

**Мониторинг температуры:**

- Пользователи могут просматривать текущую температуру в своих домах через веб-интерфейс.
- Система поддерживает получение данных о температуре через синхронные запросы от сервера к датчикам.
- Данные отображаются в реальном времени, но только для уже подключенных устройств.

### 2. Анализ архитектуры монолитного приложения

Характеристика текущего приложения:
- Язык программирования: Go
- База данных: PostgreSQL
- Архитектура: Монолитная – все компоненты (обработка запросов, бизнес-логика, доступ к данным) объединены в одном приложении.
- Взаимодействие: Синхронное, последовательная обработка запросов.
- Масштабируемость: Ограничена, так как нельзя масштабировать отдельные компоненты.
- Развертывание: Требует полной остановки приложения для обновлений.
- Подключение устройств: Только через специалистов, самостоятельное подключение пользователем невозможно.
- Текущий охват: 100 веб-клиентов и 100 модулей управления отоплением.

### 3. Определение доменов и границы контекстов

1. Домен «Управление устройствами»
   - Контекст: Самостоятельное подключение устройств пользователем, поддержка протоколов партнёров.
   - Процессы, реализуемые доменом: Регистрация, аутентификация, хранение метаданных устройств, управление состоянием подключения.
   - Сущности: Устройство, датчик, реле, ворота, камера.

2. Домен «Телеметрия и мониторинг»
   - Контекст: Мониторинг температуры, наблюдение за домом, история изменений.
   - Процессы, реализуемые доменом: Сбор, хранение, агрегация и предоставление данных с датчиков (температура, состояние устройств).
   - Сущности: Показания датчиков, метрики, журналы событий.

3. Домен «Управление сценариями»
   	- Контекст: Программирование системы пользователем под свои нужды.
	- Процессы, реализуемые доменом: Создание, выполнение и управление автоматическими сценариями (правилами) — например, «включить свет при движении».
 	- Сущности: Сценарий, правило, триггер, действие.

4. Домен «Пользователи и безопасность»
   	- Контекст: SaaS-модель, мультитенантность, самообслуживание.
	- Процессы, реализуемые доменом: Управление пользователями, аутентификация, авторизация, роли, разрешения.
	- Сущности: Пользователь, роль, сессия, токен.

5. Домен «Уведомления и события»
   	- Контекст: Оповещения о событиях, срабатывании сценариев, аномалиях.
	- Процессы, реализуемые доменом: Отправка уведомлений пользователям (email, push), обработка системных событий.
	- Сущности: Уведомление, шаблон, канал доставки.

6. Домен «Платежи и подписки»
   - Контекст: Продажа модульных комплектов, SaaS-модель.
   - Процессы, реализуемые доменом: Управление тарифами, подписками, оплатой, учётом использования.
   - Сущности: Подписка, транзакция, тариф, счёт.

7. Домен «Интеграция с внешними устройствами»
   - Контекст: Расширяемость экосистемы.
   - Процессы, реализуемые доменом: Поддержка протоколов взаимодействия с устройствами партнёров (Zigbee, Z-Wave, MQTT и др.).
   - Сущности: Адаптер протокола, драйвер устройства.

Ключевые домены для MVP:
Если говорить о MVP с управлением:
		- Отоплением
		- Освещением
		- Наблюдением
		- Воротами
Поэтому для MVP разработкуначать с доменов:
		- Управление устройствами
		- Телеметрия
		- Управление сценариями
		- Пользователи и безопасность
### **4. Проблемы монолитного решения**

1. Ограниченная масштабируемость
	- Только вертикальное масштабирование. Нельзя масштабировать отдельные модули (например, только управление освещением при высокой нагрузке).
	- Ресурсы расходуются неэффективно: даже если нагрузка только на одном функциональном модуле, приходится масштабировать весь монолит.
	- Не поддерживается геораспределение, что критично для обслуживания нескольких регионов, невозможность масштабировать географию продаж.

2. Сложность внедрения новых функций
	- Жёсткая связанность компонентов: изменение в одной части системы (например, добавление поддержки новых датчиков) требует пересборки и переразвёртывания всего приложения.
	- Долгий цикл разработки: из-за необходимости тестировать весь монолит даже при небольших изменениях.
	- Невозможно независимое развёртывание новых модулей (умный свет, ворота, камеры).

3. Низкая отказоустойчивость
	- Единая точка отказа: сбой в одном модуле (например, в работе с БД) приводит к падению всей системы.
	- Нет изоляции отказов: ошибка в новом модуле управления освещением может «повалить» и управление отоплением.
	- Сложное восстановление: при падении системы перезапускается всё приложение целиком, а не только проблемный компонент.

4. Проблемы с производительностью
	- Синхронная обработка запросов создаёт узкие места: долгий запрос к датчику блокирует обработку других запросов.
	- Отсутствие асинхронности: нет возможности фоновой обработки задач (например, сбор телеметрии, генерация отчётов).
	- Нет кэширования и оптимизированных хранилищ под разные типы данных.
	

### 5. Визуализация контекста системы — диаграмма С4

Добавьте сюда диаграмму контекста в модели C4.

Чтобы добавить ссылку в файл Readme.md, нужно использовать синтаксис Markdown. Это делают так:

```markdown
[Диаграмма контекста С4 монолита "Теплый дом"](https://www.plantuml.com/plantuml/png/jLJDIXjH5DxdAMwpIOJeK73bobQaBLX8gwKRGSPaq87v2JCJMdV-K5k8r5A4RjhQr1SO9XaROn9VuTmtwldkrOnfHiYcY3evEVTyF-Uxinn6PX0rNQVmoFPgJhDkYTqeQeHBXX6OxnPsx6YtkTqChQ3cUv7bHGirtpKQZkdXp7mOrHrxsrXPdA-YzERbck6QOMG5NDfQuHEcxLd1GWFpI-8_8BWDtyCPXI0YEt8iGMVKevWwSf__3lWd5jId4Gtb8QNKbmVyFnPbQYyAkk2c4ILS7yeJyNXEp0jgKq_rfEfeaNxpqmrTonPT2Ufo-w28jW56ykyrjn-gfrJG3S_hSMb4bvZ9pscXQWMXxvNDkcjx5PxtXOSC-9kXnW5D7Z2I-qLyUw5cn9GmuYcbyb38WumG3xwg23uQew8ReLOrCiiHp-xpoS_oeyQNfyMvEmOm6H3dO2rFw8-I3l0KdnNyLoOVKnSLddaiv163Gbo7HhVgATxqScZI22hWr9rog0HxPWsky4be6ToLQjtoIExBeunZsB1gjkdOOLGfKblr7Ik1rze4JP9taJBzoplXNR38rKyfa-q2YIizbODaYgeogR2-nxboowxl-Gw74a3VPzonAEMafAIujdNgbiGploAgVXZj1DRwgxLDC83heCGkPtQEP-OPoc8qfvwEZ8fHPyfLU2fRB4fhsbfz7vL3omjzjV-pp8xyVRHq4dN443aLlmG9I6x5AIDoStJA7DkAx-1-Blly8URMM1kU3_mFvS1op3rG-QNpcECJJyhpISTa70N5mw14jYPBEMUHX9BWiw0A3OXKEhQ00GklBUVsBjtFocmtWiqvXwUu5dfJN4_0vk7BmTEVRsnKrustgq-P7k3LooJ3pt-jz024uHfyuaKYMypa1yNPZh6X1tOA_NaiiiiwIGNhwcGq84BfmmrIIy0cfz8LmniuerQj-vX75xJw0ek32hp6fprHT38AMTQ_KQpt8Z_DRBJnfFf2bE4MWtrG78UWFiT0apZZWFxaYLFhx-EYAFUxoku46iOfp8d7B0IMBQ_UT9s_)
```


```markdown
[Диаграмма контекста микросервис "Умный дом" - MVP](https://www.plantuml.com/plantuml/uml/hLTTRnD75xxthvWtDmIzEBvvhxZYuqrfLD8Qx8A48aMRU76iskl8NbEuomSGgb024LLHcn8eKkygRHmxB4wo_Wkp_wZdERDJnhjZJ2GKnVREdZaVppddcJENshxJyZkDUku_DM-vtgcuOiNtLziN2uMMy-DCjUQlT9OwRRUrtFHyr_DdbfkDmchTyR0dt_OhzTfIuUh_yaKiVJz_lT1mshxRegN5gxJ3gNbkQsOL-lsQNtU5V2b36SZ3U8q-J-Wl478W8taQRyY-_21FPPzUHR9xKSY_yIPUbs6yZhSiF2NVoHEItfOVXEoHuCdKZ91xf7H0-o2D_Lti8Lr1l2FeOor-94Ee8A6DU1EsnNJyc9wFwR_x2R7pUJ5_gvZBdPFxPEiuVYhVqpTsRF3kC5TqM-scDms8BeWfcnYUf_1kbpO4igVCnez5l8bGIF1GHXmC8bpmv5k69aBKzBcZTK2KK1dAY6G7jF8itf81QTeWJH1cu3xW5qnV40o0MWy8eAuMNkVGyNADbaD1wrsOWddw3idvW4Je5UaOGgkHhv1WscGz_2e0reYBHJa3K0EZfaaEWLFOLw_YPq9s1VtgamW71_kqnaidx5z8MBpJLaXe1vi9WOYT7K8Xf21Frc6vIn8_cQkq4h5N6-n-o2iXvo020l80EE9PxvawhrFgL1er3tbxBdlO2maD9sTc3LdDUYZCaaL6A8oC2WD7v99jdt9SNHBVO-z2PGNVEW_8JPSB83JBWdK-fOr_M3EEw5dFmtXRo53Nq13EaFNmOyohH0QUhcHqn8Rgo2bUIRj0V-n2jluGnAkagGz7euq8yuWVUwf75FBibE3DMEbfwzJf9xgq0Cy03OPaaeKtq6v35B2hFWljld6RqXlTG9UBStdu0m5ICYH4gEV_6c6851ELBbykvKeFYEqQq-voiys_09EDjivbQEiSGxf7nWX0aaC-kIc1cfbJJGS0ZhIIr_88kTN09UNIKvfW0s0EO4Fr2BFPjXs7inFdWdUjAVuhld7k_k3GTx7PzgijjtJZEZtCUNVhdVl_lxBW4MdcgTesITM0gKHrs_4qBUbswNviO7jFVsPBdutGB5IqGqe3vH-Y1IJCPmTWDgpHkqDRtvnNPRBBFV8a2SNCFDgj1rxWKqGLo-BiVNzwnNNycbT52VncPWhyfXbb8a1SGNzotlHn4I3LbDUN_6kBGmavPhO1qK1baiNUA2oeDj8caVy16ps6Vvy-bMG_TInf7maV8-XwhRfojgWj5O2eTJD1Whc0jYASNLhI7lnxGZCQ8SaWuHpspMjNyie8qFeYeLKTtssV8QvK3MjwHLzpBsyR8fbqmY5Eov7ErHU9eTrftNDhTHhabbrxABkVEZeHmrmnZzxHKBFd4Rlr49kmefKWEj1t-aorbA_iWPU-HBGvuk7if0Hvuwp3gMJWei_4nOfxhxRCMPwW97wXb71MVoUC3hXKrRYN_CmgM3Zh0IZcckK5hrZlLFEELydJKFw0ICeuT9dp6Axd2Hz6hAUlpaQD2tY0nuOdCiHRpMC_mzGecm722OHd8kSsd5fzaIPZ04TWpE8vRnurYiAuFqx1qiPfeP1IP4kPLQFn7dFy0DIjQ1auio1qaE0wFHmdCusPmqvxnN2aI8zPFt2CCkUJEPqdi4yocZDW6b0y8J_0PBHO7ZJeccJOalj85qc-a4V9143GeNNUizNNye3ZF-3y3hYIU0S0HSrZVjnF3soK1Z5fCe0Yssra9tVJhISt5dDoOIi3d0Koq7FKvC52T-FdtqnuHqHEzSRUICBKxG9Ru4roOjjl5BRhWjK1K2u9lqw0tZ96PVE81nP8qTVbSh5Kk3rxvO8OgH5DDh1kyB6c6c6V-aU6Niu4NpHIgPYyjUGufRmk5V0hnW2amtxYc9SC1i2yBG0Sb9vnKx3TH8n9AD0DeDP6O4gYtpVv46XYghGo5CEWUJO5XFqMmTXtntBfPzuDnbm4T0b30FCLT1WQA1JauKI_q0UHkkNVd2sLeKKK19G30O2LFXu2FJg2z4Bjb8uWuTMnnOjnKe5cO6NCbECh81c9zKZ7Lx7ny-7RqRv4VFCtokK2Kaq1f1mmvpkR1sgMqzRJoZF3NhmpOgzmfrPTSjsiMJrxMMsQbpfjLyt9mDQOiYR4bXgEZ55egCzbdblwoMs9wzeskVmg0tscSRcKIqKLtdVVtfhxIbN5AqkBmDbXfgEGpt1ycQIZi8mol_15NjtukCktqAYMDjNjGl8zBfGZ25e7blXHPc11oqmQH40iB7rAz1EIvKrjsoW5LGM5E_dRpZstCD-iB7NQrckOSRh7ssaWx9E9_RY7wT8yPOYwEbiOribYh1vxSLc6XibJWKasfFAIwrKwZVe_)
```

# Задание 2. Проектирование микросервисной архитектуры

В этом задании вам нужно предоставить только диаграммы в модели C4. Мы не просим вас отдельно описывать получившиеся микросервисы и то, как вы определили взаимодействия между компонентами To-Be системы. Если вы правильно подготовите диаграммы C4, они и так это покажут.

**Диаграмма контейнеров (Containers)**

Добавьте диаграмму.

**Диаграмма компонентов (Components)**

Добавьте диаграмму для каждого из выделенных микросервисов.

**Диаграмма кода (Code)**

Добавьте одну диаграмму или несколько.

# Задание 3. Разработка ER-диаграммы

Добавьте сюда ER-диаграмму. Она должна отражать ключевые сущности системы, их атрибуты и тип связей между ними.

# Задание 4. Создание и документирование API

### 1. Тип API

Укажите, какой тип API вы будете использовать для взаимодействия микросервисов. Объясните своё решение.

### 2. Документация API

Здесь приложите ссылки на документацию API для микросервисов, которые вы спроектировали в первой части проектной работы. Для документирования используйте Swagger/OpenAPI или AsyncAPI.

# Задание 5. Работа с docker и docker-compose

Перейдите в apps.

Там находится приложение-монолит для работы с датчиками температуры. В README.md описано как запустить решение.

Вам нужно:

1) сделать простое приложение temperature-api на любом удобном для вас языке программирования, которое при запросе /temperature?location= будет отдавать рандомное значение температуры.

Locations - название комнаты, sensorId - идентификатор названия комнаты

```
	// If no location is provided, use a default based on sensor ID
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// If no sensor ID is provided, generate one based on location
	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}
```

2) Приложение следует упаковать в Docker и добавить в docker-compose. Порт по умолчанию должен быть 8081

3) Кроме того для smart_home приложения требуется база данных - добавьте в docker-compose файл настройки для запуска postgres с указанием скрипта инициализации ./smart_home/init.sql

Для проверки можно использовать Postman коллекцию smarthome-api.postman_collection.json и вызвать:

- Create Sensor
- Get All Sensors

Должно при каждом вызове отображаться разное значение температуры

Ревьюер будет проверять точно так же.


# **Задание 6. Разработка MVP**

Необходимо создать новые микросервисы и обеспечить их интеграции с существующим монолитом для плавного перехода к микросервисной архитектуре. 

### **Что нужно сделать**

1. Создайте новые микросервисы для управления телеметрией и устройствами (с простейшей логикой), которые будут интегрированы с существующим монолитным приложением. Каждый микросервис на своем ООП языке.
2. Обеспечьте взаимодействие между микросервисами и монолитом (при желании с помощью брокера сообщений), чтобы постепенно перенести функциональность из монолита в микросервисы. 

В результате у вас должны быть созданы Dockerfiles и docker-compose для запуска микросервисов. 
