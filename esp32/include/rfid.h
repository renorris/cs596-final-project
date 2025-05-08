struct RFIDResult {
    int err;
    bool isNew;
    byte uuid[16];
};

RFIDResult doRFIDLogic();
