#include <stddef.h>
#include <stdint.h>

typedef void (*task_callback)(void *context);

typedef struct {
    uint64_t deadline_ms;
    task_callback callback;
    void *context;
    int priority;
} scheduled_task;

static int compare_tasks(const scheduled_task *left, const scheduled_task *right) {
    if (left->deadline_ms != right->deadline_ms) {
        return left->deadline_ms < right->deadline_ms ? -1 : 1;
    }
    if (left->priority != right->priority) {
        return left->priority > right->priority ? -1 : 1;
    }
    return 0;
}

void sort_tasks(scheduled_task *tasks, size_t count) {
    for (size_t index = 1; index < count; index++) {
        scheduled_task current = tasks[index];
        size_t position = index;
        while (position > 0 && compare_tasks(&current, &tasks[position - 1]) < 0) {
            tasks[position] = tasks[position - 1];
            position--;
        }
        tasks[position] = current;
    }
}

size_t run_ready_tasks(scheduled_task *tasks, size_t count, uint64_t now_ms) {
    size_t completed = 0;
    while (completed < count && tasks[completed].deadline_ms <= now_ms) {
        if (tasks[completed].callback != NULL) {
            tasks[completed].callback(tasks[completed].context);
        }
        completed++;
    }
    return completed;
}
