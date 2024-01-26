#include <stdio.h>
#include <malloc/_malloc.h>
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>

#define EXIT "exit"

// 命令行提示输入
void print_prompt() {
    printf("neptune-db > ");
}

// 输入缓冲区
struct InputBuffer_t {
    // 字符串缓冲区
    char* buffer;
    // 缓冲区大小
    size_t buffer_length;
    // 读取字符串的长度
    ssize_t input_length;
};

// 定义输入缓冲区别名
typedef struct InputBuffer_t InputBuffer;

// 定义构造函数
InputBuffer* new_input_buffer() {
    // 1. 实例化输入缓冲区实例
    InputBuffer* input_buffer = malloc(sizeof(InputBuffer));
    // 2. 初始化缓冲区
    input_buffer->buffer = NULL;
    // 3. 初始化缓冲区大小
    input_buffer->buffer_length = 0;
    // 4. 初始化读取字符串长度大小
    input_buffer->input_length = 0;
}

void read_input(InputBuffer* input_buffer) {
    // 1. 读取控制台的输入: 缓冲区、缓冲区大小、数据源
    ssize_t input_len = getline(&(input_buffer->buffer), &(input_buffer->buffer_length), stdin);
    // 2. 判断读取的内容是否合法
    if (input_len <= 0) {
        print_prompt();
        return;
    }
    // 3. 过滤换行符
    input_buffer->input_length = input_len - 1;
    input_buffer->buffer[input_len - 1] = 0;
}


int main() {
    // 1. 初始化输入缓冲区
    InputBuffer* input_buffer = new_input_buffer();
    // 2. 持续循环读取控制台输入
    while (true) {
        // 2.1 控制台提示输入
        print_prompt();
        // 2.2 读取控制台输入
        read_input(input_buffer);

        // 2.3 判断是否退出
        if (strcmp(input_buffer->buffer, EXIT) == 0){
            exit(EXIT_SUCCESS);
        } else {
            printf("unrecognized command '%s'.\n", input_buffer->buffer);
        }
    }
}
