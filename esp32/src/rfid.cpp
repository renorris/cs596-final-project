#include <Arduino.h>
#include <SPI.h>
#include <MFRC522.h>

#include "rfid.h"
#include "bootloader_random.h"
#include "esp_random.h"

#include "pins.h"
#include "credentials.h"

// Default MIFARE key (all bits set)
MFRC522::MIFARE_Key mifareDefaultKey = {
    .keyByte = { 
        0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF
    }
};

// Private MIFARE key
MFRC522::MIFARE_Key mifarePrivateKey = {
    .keyByte = { MIFARE_PRIVATE_KEY }
};

RFIDResult doRFIDLogic() {

    RFIDResult result = RFIDResult{0};

    SPI.end();
    SPI.begin();
    MFRC522 mfrc522(RC522_SS_PIN, RC522_RST_PIN);

    // Attempt to read the card with the default key.
    // If it works, its a new card, and a new random
    // key should be written to it.

    mfrc522.PCD_Init();
    mfrc522.PCD_DumpVersionToSerial();

    // Check if there's a card present
    if (!mfrc522.PICC_IsNewCardPresent()) {
        Serial.println("No card present");
        result.err = -1;
        return result;
    }
    
    // Attempt to read the card UID
    if (!mfrc522.PICC_ReadCardSerial()) {
        Serial.println("Unable to read card");
        result.err = -1;
        return result;
    }

    const byte dataBlockAddr = 0x02;
    const byte trailerBlockAddr = 0x03;

    // Attempt to authenticate with default key
    MFRC522::StatusCode status = mfrc522.PCD_Authenticate(
        MFRC522::PICC_CMD_MF_AUTH_KEY_A, 
        trailerBlockAddr, 
        &mifareDefaultKey, 
        &(mfrc522.uid)
    );

    // If it worked, write a random value to the card
    if (status == MFRC522::STATUS_OK) {
        // Generate cryptographically random UUID
        bootloader_random_enable();
        esp_fill_random(result.uuid, 16);

        // Write random UUID to data block
        status = mfrc522.MIFARE_Write(dataBlockAddr, result.uuid, 16);
        if (status != MFRC522::STATUS_OK) {
            Serial.println("Unable to write new random UUID to card");
            result.err = -1;
            return result;
        }

        Serial.println("Wrote new random UUID to card:");
        for (int i = 0; i < 16; i++) {
            Serial.printf("%02x", result.uuid[i]);
        }
        Serial.print("\n");

        // Build new trailer block
        byte trailerBlockData[16] = {
            MIFARE_PRIVATE_KEY, // Key A
            0xFF, 0x07, 0x80,   // Access bits
            0xFF,               // User byte
            MIFARE_PRIVATE_KEY  // Key B
        };

        Serial.println("Attempting to write trailer block:");
        for (int i = 0; i < 16; i++) {
            Serial.printf("%02x", trailerBlockData[i]);
        }
        Serial.print("\n");

        // Write the trailer block to the card
        status = mfrc522.MIFARE_Write(trailerBlockAddr, trailerBlockData, 16);
        if (status != MFRC522::STATUS_OK) {
            Serial.println("Unable to write new security block to card:");
            Serial.println(MFRC522::GetStatusCodeName(status));
            result.err = -1;
            return result;
        }

        Serial.println("Wrote new security block to card");

        result.isNew = true;
        return result;
    }

    Serial.println("Unable to authenticate using default key:");
    Serial.println(MFRC522::GetStatusCodeName(status));
    Serial.println("Trying with private key...");

    // Check if there's a card present
    if (!mfrc522.PICC_IsNewCardPresent()) {
        Serial.println("No card present");
        result.err = -1;
        return result;
    }
    
    // Attempt to read the card UID
    if (!mfrc522.PICC_ReadCardSerial()) {
        Serial.println("Unable to read card");
        result.err = -1;
        return result;
    }

    // Attempt to authenticate with private key
    status = mfrc522.PCD_Authenticate(
        MFRC522::PICC_CMD_MF_AUTH_KEY_A, 
        trailerBlockAddr, 
        &mifarePrivateKey, 
        &(mfrc522.uid)
    );
    if (status != MFRC522::STATUS_OK) {
        Serial.println("Unable to authenticate card with private key:");
        Serial.println(MFRC522::GetStatusCodeName(status));
        result.err = -1;
        return result;
    }

    byte bufLen = 18; 
    byte buf[bufLen];
    status = mfrc522.MIFARE_Read(dataBlockAddr, buf, &bufLen);
    if (status != MFRC522::STATUS_OK) {
        Serial.println("Unable to read data block with private key:");
        Serial.println(MFRC522::GetStatusCodeName(status));
        result.err = -1;
        return result;
    }

    Serial.println("Successfully authenticated with private key");

    // Copy UUID into result (last 2 bytes are CRC and are verified by the library)
    for (int i = 0; i < 16; i++) {
        result.uuid[i] = buf[i];
    }

    result.isNew = false;
    return result;
}
