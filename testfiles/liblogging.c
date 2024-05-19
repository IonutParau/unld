// This is an implementation of a super basic logging library
// This is to test global state
// Yes, global C library state is evil but this is to test the functionality of unld

#include "liblogging.h"
#include "stdio.h"

struct LogState {
    char *header;
    char **levelInfo;
} logState;

void log_init(char *header, char **levelInfo) {
    logState.header = header;
    logState.levelInfo = levelInfo;
}

void log_raw(char *message, int severity) {
    if(severity >= LOG_NAMEC) {
        fprintf(stderr, "Severity out of range (%d) for message %s\n", severity, message);
        return;
    }

    printf("%s%s%s\n", logState.header, logState.levelInfo[severity], message);
}

void log_info(char *message) {
    log_raw(message, LOG_INFO);
}
void log_debug(char *message) {
    log_raw(message, LOG_DEBUG);
}
void log_error(char *message) {
    log_raw(message, LOG_ERR);
}
void log_warn(char *message) {
    log_raw(message, LOG_WARN);
}
