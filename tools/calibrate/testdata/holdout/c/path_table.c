#include <stdbool.h>
#include <stddef.h>
#include <stdio.h>
#include <string.h>

#define PATH_TABLE_SIZE 32
#define PATH_LIMIT 96

typedef struct {
    char key[PATH_LIMIT];
    unsigned visits;
    bool occupied;
} path_entry;

static unsigned path_hash(const char *path) {
    unsigned hash = 2166136261u;
    for (const unsigned char *cursor = (const unsigned char *)path; *cursor; cursor++) {
        hash ^= *cursor;
        hash *= 16777619u;
    }
    return hash;
}

bool record_path(path_entry *table, const char *path) {
    if (strlen(path) >= PATH_LIMIT) {
        return false;
    }
    unsigned start = path_hash(path) % PATH_TABLE_SIZE;
    for (unsigned offset = 0; offset < PATH_TABLE_SIZE; offset++) {
        path_entry *entry = &table[(start + offset) % PATH_TABLE_SIZE];
        if (!entry->occupied) {
            snprintf(entry->key, sizeof(entry->key), "%s", path);
            entry->visits = 1;
            entry->occupied = true;
            return true;
        }
        if (strcmp(entry->key, path) == 0) {
            entry->visits++;
            return true;
        }
    }
    return false;
}
