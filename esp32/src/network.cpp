#include <HTTPClient.h>

#include "network.h"
#include "esp_wpa2.h"
#include "credentials.h"
#include "rfid.h"

RTC_DATA_ATTR HTTPClient client;

void beginConnectToWiFi() {
    // Set WiFi to Station mode and disconnect from any previous AP
    WiFi.mode(WIFI_STA);
    WiFi.disconnect();

    #ifdef EAP_ID
    // Configure WPA2 Enterprise credentials
    esp_wifi_sta_wpa2_ent_set_identity((uint8_t *)EAP_ID, strlen(EAP_ID));
    esp_wifi_sta_wpa2_ent_set_username((uint8_t *)EAP_USERNAME, strlen(EAP_USERNAME));
    esp_wifi_sta_wpa2_ent_set_password((uint8_t *)EAP_PASSWORD, strlen(EAP_PASSWORD));
    esp_wifi_sta_wpa2_ent_enable();
    WiFi.begin(WIFI_SSID);
    #else
    // Normal WPA2 Personal
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
    #endif
}

void requestCreateNewCard(byte uuid[]) {
    
    Serial.println("begin new card request");
    
    // Spin up an HTTP client
    while (!client.begin(NEW_CARD_URL)) {
        Serial.println("new card client failed");
    }

    Serial.println("HTTP begin");

    // Set configured basic auth
    client.addHeader("Authorization", BASIC_AUTH);
    client.addHeader("Content-Type", "application/json");

    // Format UUID as hex string
    char uuidBuf[33];
    for (int i = 0; i < 16; i++) {
        char buf[3]; // Buffer for "%02x\0"
        sprintf(buf, "%02x", uuid[i]);
        uuidBuf[2*i] = buf[0];
        uuidBuf[2*i+1] = buf[1];
    }
    uuidBuf[32] = '\0';

    // Format JSON request body
    char jsonBuf[128];
    sprintf(jsonBuf, "{\"uuid\":\"%s\"}", uuidBuf);
    Serial.printf("%s\n", jsonBuf);
    String requestBody = String(jsonBuf);

    // Do the request
    int responseCode = client.POST(requestBody);

    Serial.println("POST done");

    client.end();

    Serial.println("HTTP end");

    return;
}

// Check if a given UUID has access to open the lockbox
bool requestLockboxAccess(byte uuid[]) {
    Serial.println("begin request lockbox access");
    
    // Spin up an HTTP client
    client.setReuse(true);
    while (!client.begin(USE_CARD_URL)) {
        Serial.println("use card client failed");
    }

    Serial.println("HTTP client begin");

    // Set configured basic auth
    client.addHeader("Authorization", BASIC_AUTH);
    client.addHeader("Content-Type", "application/json");

    // Format UUID as hex string
    char uuidBuf[33];
    for (int i = 0; i < 16; i++) {
        char buf[3]; // Buffer for "%02x\0"
        sprintf(buf, "%02x", uuid[i]);
        uuidBuf[2*i] = buf[0];
        uuidBuf[2*i+1] = buf[1];
    }
    uuidBuf[32] = '\0';

    // Format JSON request body
    char jsonBuf[128];
    sprintf(jsonBuf, "{\"uuid\":\"%s\"}", uuidBuf);
    Serial.printf("%s\n", jsonBuf);
    String requestBody = String(jsonBuf);

    // Do the request
    int responseCode = client.POST(requestBody);

    Serial.println("HTTP done post");

    client.end();

    Serial.println("HTTP end");

    return responseCode == HTTP_CODE_NO_CONTENT;
}
