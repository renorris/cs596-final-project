#include <WiFi.h>

// Asynchronously begin connecting to WiFi
void beginConnectToWiFi();

bool requestLockboxAccess(byte uuid[]);

void requestCreateNewCard(byte uuid[]);
