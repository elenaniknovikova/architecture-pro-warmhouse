const express = require('express');
const app = express();
const PORT = process.env.PORT || 8081;

// Функция для генерации случайной температуры от -10 до +40
const generateRandomTemperature = () => {
    return (Math.random() * 50 - 10).toFixed(1);
};

app.get('/temperature', (req, res) => {
    let { location, sensorId } = req.query;

    // --- ПОЛНАЯ ЛОГИКА ИЗ ЗАДАНИЯ ---
    // Если location не передан, определяем по sensorId
    if (!location && sensorId) {
        switch (sensorId) {
            case "1":
                location = "Living Room";
                break;
            case "2":
                location = "Bedroom";
                break;
            case "3":
                location = "Kitchen";
                break;
            default:
                location = "Unknown";
        }
    }

    // Если sensorId не передан, генерируем на основе location
    if (!sensorId && location) {
        switch (location) {
            case "Living Room":
                sensorId = "1";
                break;
            case "Bedroom":
                sensorId = "2";
                break;
            case "Kitchen":
                sensorId = "3";
                break;
            default:
                sensorId = "0";
        }
    }

    // Если после всех проверок нет данных, ставим значения по умолчанию
    if (!location && !sensorId) {
        location = "Unknown";
        sensorId = "0";
    }

    // Генерируем случайную температуру
    const temperature = generateRandomTemperature();

    // Формируем ответ
    const response = {
        sensorId: sensorId,
        location: location,
        temperature: parseFloat(temperature),
        unit: 'celsius',
        timestamp: new Date().toISOString()
    };

    console.log(`[${new Date().toISOString()}] GET /temperature - ${location} (${sensorId}): ${temperature}°C`);
    res.json(response);
});

app.get('/health', (req, res) => {
    res.status(200).send('OK');
});

app.listen(PORT, () => {
    console.log(`Temperature API running on port ${PORT}`);
});