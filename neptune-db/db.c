#include <stdio.h>
#include <malloc/_malloc.h>
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>

#define EXIT ".exit"
#define INSERT "insert"
#define SELECT "select"

#define size_of_attribute(Struct, Attribute) sizeof(((Struct*)0)->Attribute);

#define COLUMN_EMAIL_SIZE 255
#define COLUMN_USERNAME_SIZE 32
#define TABLE_MAX_PAGES 100


// 硬编码行
typedef struct {
    // 主键
    uint32_t id;
    // 内容
    char username[COLUMN_USERNAME_SIZE + 1];
    char email[COLUMN_EMAIL_SIZE + 1];
} Row;

const uint32_t ID_SIZE = size_of_attribute(Row, id);
const uint32_t USERNAME_SIZE = size_of_attribute(Row, username);
const uint32_t EMAIL_SIZE = size_of_attribute(Row, email)
const uint32_t ID_OFFSET = 0;
const uint32_t USERNAME_OFFSET = ID_OFFSET + ID_SIZE;
const uint32_t EMAIL_OFFSET = USERNAME_OFFSET + USERNAME_SIZE;
const uint32_t ROW_SIZE = ID_SIZE + USERNAME_SIZE + USERNAME_OFFSET;

// 打印查询的内容
void print_row(Row* row) {
    printf("(%d, %s, %s)\n", row->id, row->username, row->email);
}

// 序列化每行内容
void serialize_row(Row *source, void *destination) {
    // 指针运算: 将内容拷贝到指定的内存地址
    memcpy(destination + ID_OFFSET, &(source->id), ID_SIZE);
    memcpy(destination + USERNAME_OFFSET, &(source->username), USERNAME_SIZE);
    memcpy(destination + EMAIL_OFFSET, &(source->email), EMAIL_SIZE);
}

// 反序列化每行内容
void deserialize_row(void *source, Row *destination) {
    memcpy(&(destination->id), source + ID_OFFSET, ID_SIZE);
    memcpy(&(destination->username), source + USERNAME_OFFSET, USERNAME_SIZE);
    memcpy(&(destination->email), source + EMAIL_OFFSET, EMAIL_SIZE);
}

// 表结构
typedef struct {
    uint32_t num_rows;
    void *pages[TABLE_MAX_PAGES];
} Table;

// 每页大小 4KB
const uint32_t PAGE_SIZE = 4096;
// 每页的行数
const uint32_t ROWS_PER_PAGE = PAGE_SIZE / ROW_SIZE;
// 每张表的行数
const uint32_t TABLE_MAX_ROWS = ROWS_PER_PAGE * TABLE_MAX_PAGES;

// 初始化表
Table* new_table(){
    // 1. 分配表内存
    Table* table = (Table*)malloc(sizeof(Table));
    // 2. 初始化表的行号为 0
    table->num_rows = 0;
    // 3. 初始化页内存
    for (uint32_t index = 0; index < TABLE_MAX_PAGES; index++) {
        table->pages[index] = NULL;
    }
    return table;
}

// 释放表内存
void free_table(Table* table) {
    // 如果页内存为空, 那么就需要释放
    // TODO 如果某个页中还有数据, 那不会造成内存泄露吗？
    for (uint32_t index = 0; table->pages[index]; index++) {
        free(table->pages[index]);
    }
    free(table);
}

// 读取表中的行记录
void* row_slot(Table *table, uint32_t row_num) {
    // 1. 计算行所在的页
    uint32_t page_num = row_num / ROWS_PER_PAGE;
    // 2. 获取对应的页
    void *page = table->pages[page_num];
    // 3. 判断页是否为空, 如果为空就分配内存
    if (page == NULL) {
        page = table->pages[page_num] = malloc(PAGE_SIZE);
    }
    // 4. 计算行在页中的偏移量
    uint32_t row_offset = row_num % ROWS_PER_PAGE;
    // 5. 行在页中的偏移量换算成字节偏移量
    uint32_t byte_offset = row_offset * ROW_SIZE;
    // 6. 指针运算
    return page + byte_offset;
}

// 输入缓冲区
struct InputBuffer_t {
    // 字符串缓冲区
    char *buffer;
    // 缓冲区大小
    size_t buffer_length;
    // 读取字符串的长度
    ssize_t input_length;
};

// 定义输入缓冲区别名
typedef struct InputBuffer_t InputBuffer;

// 定义构造函数
InputBuffer *new_input_buffer() {
    // 1. 实例化输入缓冲区实例
    InputBuffer *input_buffer = malloc(sizeof(InputBuffer));
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

void read_input(InputBuffer *input_buffer) {
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

void close_input_buffer(InputBuffer *input_buffer) {
    // 释放真正缓冲区占用的内存
    free(input_buffer->buffer);
    // 释放缓冲区封装类占用的内存
    free(input_buffer);
}

// 元命令类型
typedef enum {
    META_COMMAND_SUCCESS,
    META_COMMAND_UNRECOGNIZED_COMMAND
} MetaCommandResult;

// 初始化 SQL 结果
typedef enum {
    PREPARE_SUCCESS,
    PREPARE_STRING_TOO_LONG,
    PREPARE_SYNTAX_ERROR,
    PREPARE_NEGATIVE_ID,
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
    Row row_to_insert; // 仅用于插入语句
} Statement;

typedef enum {
    EXECUTE_SUCCESS,
    EXECUTE_TABLE_FULL
}ExecuteResult;

// 判断元命令类型
MetaCommandResult do_meta_command(InputBuffer *input_buffer, Table* table) {
    // 判断是否为退出的元命令
    if (strcmp(input_buffer->buffer, EXIT) == 0) {
        close_input_buffer(input_buffer);
        free_table(table);
        exit(EXIT_SUCCESS);
        // return META_COMMAND_SUCCESS;
    }
    return META_COMMAND_UNRECOGNIZED_COMMAND;
}

// 初始化插入 SQL 语句
PrepareResult prepare_insert(InputBuffer* input_buffer, Statement* statement) {
    // 1. 指定 SQL 语句类型
    statement->type = STATEMENT_INSERT;
    // 2. 读取缓冲区中的关键字
    char* keyword = strtok(input_buffer->buffer, " ");
    char* id_string = strtok(NULL, " ");
    char* username = strtok(NULL, " ");
    char* email = strtok(NULL, " ");
    // 3. 判断字段是否为空
    if (id_string == NULL || username == NULL || email == NULL) {
        return PREPARE_SYNTAX_ERROR;
    }
    // 4. 转换 id
    int id = atoi(id_string);
    if (id < 0) {
        return PREPARE_NEGATIVE_ID;
    }
    // 5. 判断 username 是否超出限制
    if (strlen(username) > COLUMN_USERNAME_SIZE) {
        return PREPARE_STRING_TOO_LONG;
    }
    if (strlen(email) > COLUMN_EMAIL_SIZE) {
        return PREPARE_STRING_TOO_LONG;
    }
    statement->row_to_insert.id = id;
    // TODO 为什么不直接赋值呢?
    strcpy(statement->row_to_insert.username, username);
    strcpy(statement->row_to_insert.email, email);
    return PREPARE_SUCCESS;
}

// 初始化 SQL 语句
PrepareResult prepare_statement(InputBuffer *input_buffer, Statement *statement) {
    // 1. 判断是否为插入语句
    if (strncmp(input_buffer->buffer, INSERT, 6) == 0) {
        return prepare_insert(input_buffer, statement);
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

// 执行 SQL 插入语句
ExecuteResult execute_insert(Statement* statement, Table* table) {
    // 1. 判断表的行号是否超过限制
    if (table->num_rows >= TABLE_MAX_ROWS) {
        return EXECUTE_TABLE_FULL;
    }
    // 2. 获取需要插入的行记录
    Row* row_to_insert = &(statement->row_to_insert);
    // 3. 序列化行记录
    serialize_row(row_to_insert, row_slot(table, table->num_rows));
    // 4. 增加表的行记录数量
    table->num_rows += 1;
    return EXECUTE_SUCCESS;
}

// 执行 SQL 查询语句
ExecuteResult execute_select(Statement* statement, Table* table) {
    Row row;
    for (uint32_t index = 0; index < table->num_rows; index++) {
        deserialize_row(row_slot(table, index), &row);
        print_row(&row);
    }
    return EXECUTE_SUCCESS;
}

// 执行 SQL 语句
ExecuteResult execute_statement(Statement *statement, Table* table) {
    // 判断 SQL 语句类型
    switch (statement->type) {
        case STATEMENT_INSERT:
            return execute_insert(statement, table);
        case STATEMENT_SELECT:
            return execute_select(statement, table);
    }
}

// TODO 为什么存在内存泄露的问题
int main() {
    // 1. 初始化表结构
    Table* table = new_table();
    // 2. 初始化输入缓冲区
    InputBuffer *input_buffer = new_input_buffer();
    // 3. 持续循环读取控制台输入
    while (true) {
        // 3.1 控制台提示输入
        print_prompt();
        // 3.2 读取控制台输入
        read_input(input_buffer);
        // 3.3 判断是否为元命令 / 非 SQL 命令
        if (input_buffer->buffer[0] == '.') {
            switch (do_meta_command(input_buffer, table)) {
                case META_COMMAND_SUCCESS:
                    continue;
                case META_COMMAND_UNRECOGNIZED_COMMAND:
                    printf("unrecognized command '%s'.\n", input_buffer->buffer);
                    continue;
            }
        }
        // 3.4 初始化 SQL 语句
        Statement statement;
        switch (prepare_statement(input_buffer, &statement)) {
            case PREPARE_SUCCESS:
                break;
            case PREPARE_NEGATIVE_ID:
                printf("id must be positive. \n");
                continue;
            case PREPARE_STRING_TOO_LONG:
                printf("string is too long. \n");
                continue;
            case PREPARE_SYNTAX_ERROR:
                printf("syntax error. could not parse statement. \n");
                continue;
            case PREPARE_UNRECOGNIZED_STATEMENT:
                printf("unrecognized keyword at start of '%s'.\n", input_buffer->buffer);
                continue;
        }
        // 2.5 执行 SQL 语句
        switch (execute_statement(&statement, table)) {
            case EXECUTE_SUCCESS:
                printf("executed.\n");
                break;
            case EXECUTE_TABLE_FULL:
                printf("error: table full. \n");
        }
    }
}
