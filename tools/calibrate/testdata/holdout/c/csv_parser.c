#include <stdbool.h>
#include <stddef.h>
#include <string.h>

typedef struct {
    const char *start;
    size_t length;
} csv_field;

size_t split_csv_line(const char *line, csv_field *fields, size_t capacity) {
    size_t count = 0;
    const char *field_start = line;
    bool quoted = false;

    for (const char *cursor = line; ; cursor++) {
        char current = *cursor;
        if (current == '"') {
            if (quoted && cursor[1] == '"') {
                cursor++;
                continue;
            }
            quoted = !quoted;
        }
        if ((current == ',' && !quoted) || current == '\0') {
            if (count < capacity) {
                fields[count].start = field_start;
                fields[count].length = (size_t)(cursor - field_start);
            }
            count++;
            field_start = cursor + 1;
        }
        if (current == '\0') {
            break;
        }
    }
    return count;
}

bool csv_field_equals(csv_field field, const char *expected) {
    size_t expected_length = strlen(expected);
    return field.length == expected_length &&
           strncmp(field.start, expected, field.length) == 0;
}
