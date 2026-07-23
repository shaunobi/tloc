#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#define RING_CAPACITY 16

typedef struct {
    int values[RING_CAPACITY];
    size_t head;
    size_t length;
} ring_buffer;

void ring_init(ring_buffer *ring) {
    ring->head = 0;
    ring->length = 0;
}

bool ring_push(ring_buffer *ring, int value) {
    if (ring->length == RING_CAPACITY) {
        return false;
    }
    size_t slot = (ring->head + ring->length) % RING_CAPACITY;
    ring->values[slot] = value;
    ring->length++;
    return true;
}

bool ring_pop(ring_buffer *ring, int *value) {
    if (ring->length == 0 || value == NULL) {
        return false;
    }
    *value = ring->values[ring->head];
    ring->head = (ring->head + 1) % RING_CAPACITY;
    ring->length--;
    return true;
}

int64_t ring_sum(const ring_buffer *ring) {
    int64_t total = 0;
    for (size_t index = 0; index < ring->length; index++) {
        total += ring->values[(ring->head + index) % RING_CAPACITY];
    }
    return total;
}
