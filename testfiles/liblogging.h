// This is an implementation of a super basic logging library
// This is to test global state
// Yes, global C library state is evil but this is to test the functionality of unld

#define LOG_INFO 0
#define LOG_DEBUG 1
#define LOG_ERR 2
#define LOG_WARN 3

#define LOG_NAMEC 4

void log_init(char *header, char **levelInfo);

void log_raw(char *message, int severity);

void log_info(char *message);
void log_debug(char *message);
void log_error(char *message);
void log_warn(char *message);
