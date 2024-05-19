#include "liblogging.h"

int main() {
    char *header = "";
    char *logInfo[LOG_NAMEC] = {0};
    logInfo[LOG_ERR] = "Error: ";
    logInfo[LOG_WARN] = "Warning: ";
    logInfo[LOG_INFO] = "Info! ";
    logInfo[LOG_DEBUG] = "Debug! ";
    log_init(header, logInfo);

    log_info("The executable started!");
    log_debug("The executable did not get stuck");
    log_warn("The executable is about to stop");
    log_error("The executable has stopped");
}
