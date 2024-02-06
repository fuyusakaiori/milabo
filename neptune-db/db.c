#include <stdio.h>
#include <malloc/_malloc.h>
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>

#define EXIT ".exit"
#define INSERT "insert"
#define SELECT "select"

// 元命令类型
typedef enum {
    META_COMMAND_SUCCESS,
    META_COMMAND_UNRECOGNIZED_COMMAND
} MetaCommandResult;

//
typedef enum {
    PREPARE_SUCCESS,
    PREPARE_UNRECOGNIZED_STATEMENT
} PrepareResult;


// SQL 语句类型
typedef enum {
    STATEMENT_INSERT,
    STATEMENT_SELECT,
} StatementType;

// SQL 语句
typedef struct {
    StatementType type;
} Statement;

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
    return input_buffer;
}

// 命令行提示输入
void print_prompt() {
    printf("neptune-db > ");
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

void close_input_buffer(InputBuffer* input_buffer) {
    // 释放真正缓冲区占用的内存
    free(input_buffer->buffer);
    // 释放缓冲区封装类占用的内存
    free(input_buffer);
}

// 判断元命令类型
MetaCommandResult do_meta_command(InputBuffer* input_buffer) {
    // 判断是否为退出的元命令
    if (strcmp(input_buffer->buffer, EXIT) == 0) {
        close_input_buffer(input_buffer);
        exit(EXIT_SUCCESS);
        // return META_COMMAND_SUCCESS;
    }
    return META_COMMAND_UNRECOGNIZED_COMMAND;
}

// 准备 SQL 语句
PrepareResult prepare_statement(InputBuffer* input_buffer, Statement* statement) {
    // 1. 判断是否为插入语句
    if (strncmp(input_buffer->buffer, INSERT, 6) == 0) {
        // 1.1 更新 SQL 语句类型为插入类型
        statement->type = STATEMENT_INSERT;
        // 1.2 返回准备完成的结果
        return PREPARE_SUCCESS;
    }
    // 2. 判断是否为查询语句
    if (strncmp(input_buffer->buffer, SELECT, 6) == 0) {
        // 2.1 更新 SQL 语句类型为查询类型
        statement->type = STATEMENT_SELECT;
        // 2.2 返回准备完成的结果
        return PREPARE_SUCCESS;
    }
    return PREPARE_UNRECOGNIZED_STATEMENT;
}

// 执行 SQL 语句
void execute_statement(Statement* statement) {
    // 判断 SQL 语句类型
    switch (statement->type) {
        case STATEMENT_INSERT:
            printf("this is where we would do an insert\n");
            break;
        case STATEMENT_SELECT:
            printf("this is where we would do an select\n");
            break;
    }
}


int main() {
    // 1. 初始化输入缓冲区
    // TODO 为什么存在内存泄露的问题
    InputBuffer* input_buffer = new_input_buffer();
    // 2. 持续循环读取控制台输入
    while (true) {
        // 2.1 控制台提示输入
        print_prompt();
        // 2.2 读取控制台输入
        read_input(input_buffer);
        // 2.3 判断是否为元命令 / 非 SQL 命令
        if (input_buffer->buffer[0] == '.') {
            switch (do_meta_command(input_buffer)) {
                case META_COMMAND_SUCCESS:
                    continue;
                case META_COMMAND_UNRECOGNIZED_COMMAND:
                    printf("unrecognized command '%s'.\n", input_buffer->buffer);
                    continue;
            }
        }
        // 2.4 初始化 SQL 语句
        Statement statement;
        switch (prepare_statement(input_buffer, &statement)) {
            case PREPARE_SUCCESS:
                break;
            case PREPARE_UNRECOGNIZED_STATEMENT:
                printf("unrecognized keyword at start of '%s'.\n", input_buffer->buffer);
                continue;
        }
        // 2.5 执行 SQL 语句
        execute_statement(&statement);

        printf("executed.\n");
    }
}
