#include <stdio.h>
#include "utils.h"

#define MAX_SIZE 1024
#define SQUARE(x) ((x) * (x))

/* A simple point structure. */
typedef struct {
    double x;
    double y;
} Point;

struct User {
    char name[64];
    int age;
};

enum Color { RED, GREEN, BLUE };

typedef int (*callback_fn)(int, int);

/* Adds two integers and returns the sum. */
int add(int a, int b) {
    return a + b;
}

// Processes the input string.
void process(const char *input) {
    printf("%s\n", input);
}

static int helper(void) {
    return 42;
}
