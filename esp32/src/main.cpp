#include <Arduino.h>
#include <SPI.h>
#include <MFRC522.h>
#include <Deneyap_Servo.h>

#include "esp_sleep.h"
#include "driver/gpio.h"
#include "pins.h"
#include "rfid.h"
#include "network.h"

MFRC522 mfrc522(SS, RC522_RST_PIN);
MFRC522::MIFARE_Key mifareKey;

Servo servo;
RTC_DATA_ATTR volatile bool isDoorLocked = false;

void setup() {
    // Configure pins
    pinMode(BUZZER_PIN, OUTPUT);
    pinMode(MOSFET_PIN, OUTPUT);

    // Register GPIO pin wakeup for later
    esp_deep_sleep_enable_gpio_wakeup(1ULL << BUTTON_PIN, ESP_GPIO_WAKEUP_GPIO_HIGH);

    Serial.begin(CONFIG_MONITOR_BAUD);
    Serial.println("Starting...");

    // Determine if we woke up from a button press
    bool wasButtonPressed = esp_sleep_get_wakeup_cause() == ESP_SLEEP_WAKEUP_GPIO;
    if (!wasButtonPressed) {
        // Go to sleep
        esp_deep_sleep_start();
        return;
    }

    digitalWrite(MOSFET_PIN, HIGH);
    delay(100);

    // If the door is opened, close it
    if (!isDoorLocked) {
        isDoorLocked = true;
        servo.attach(SERVO_PIN);
        servo.write(89);
        delay(1000);

        digitalWrite(MOSFET_PIN, LOW);

        esp_deep_sleep_start();
        return;
    }

    servo.attach(SERVO_PIN);
    servo.write(89);

    // Start connecting to WiFi
    beginConnectToWiFi();

    RFIDResult res = doRFIDLogic();
    if (res.err != 0) {
        Serial.println("doRFIDLogic failed");
    }

    // Wait for WiFi to connect
    while (WiFi.status() != WL_CONNECTED) {
        Serial.print(".");
        delay(50);
    };

    Serial.println("Connected to WiFi.");

    bool accessGranted = false;
    if (res.isNew) {
        requestCreateNewCard(res.uuid);
    }
    else {
        accessGranted = requestLockboxAccess(res.uuid);
        requestCreateNewCard(res.uuid);
    }

    if (accessGranted) {
        servo.write(1);
        delay(1000);
        isDoorLocked = false;
    }

    WiFi.disconnect(true, false);

    // Power down peripherals
    digitalWrite(MOSFET_PIN, LOW);

    // Go to sleep
    esp_deep_sleep_start();
}

void loop() {
    return;
}
