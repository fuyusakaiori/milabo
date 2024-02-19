#include <stdio.h>
#include <malloc/_malloc.h>
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>

#define EXIT ".exit"
#define INSERT "insert"
#define SELECT "select"

#define size_of_attribute(Struct, Attribute) sizeof(((Struct*)0)->Attribute);

#define COLUMN_EMAIL_SIZE 255
#define COLUMN_USERNAME_SIZE 32
#define TABLE_MAX_PAGES 100


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
const uint32_t ROW_SIZE = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE;

// 打印查询的内容
void print_row(Row* row) {
    printf("(%d, %s, %s)\n", row->id, row->username, row->email);
}

// 序列化每行内容
void serialize_row(Row *source, void *destination) {
    // 指针运算: 将内容拷贝到指定的内存地址
    memcpy(destination + ID_OFFSET, &(source->id), ID_SIZE);
    strncpy(destination + USERNAME_OFFSET, source->username, USERNAME_SIZE);
    strncpy(destination + EMAIL_OFFSET, source->email, EMAIL_SIZE);
}

// 反序列化每行内容
void deserialize_row(void *source, Row *destination) {
    memcpy(&(destination->id), source + ID_OFFSET, ID_SIZE);
    memcpy(&(destination->username), source + USERNAME_OFFSET, USERNAME_SIZE);
    memcpy(&(destination->email), source + EMAIL_OFFSET, EMAIL_SIZE);
}

// 每页大小 4KB
const uint32_t PAGE_SIZE = 4096;
// 每页的行数
const uint32_t ROWS_PER_PAGE = PAGE_SIZE / ROW_SIZE;
// 每张表的行数
const uint32_t TABLE_MAX_ROWS = ROWS_PER_PAGE * TABLE_MAX_PAGES;

// 内存调度器
typedef struct {
    // 文件描述符: 用于判断文件的状态
    int file_descriptor;
    // 文件大小
    uint32_t file_length;
    // 管理的内存
    void* pages[TABLE_MAX_PAGES];
} Pager;

// 初始化内存调度器
Pager* pager_open(const char* filename) {
    // 1. 打开文件
    int fd = open(filename, O_RDWR | O_CREAT | S_IWUSR | S_IRUSR);
    // 2. 判断是否成功打开文件
    if (fd == -1) {
        printf("unable to open file\n");
        exit(EXIT_FAILURE);
    }
    // 3. 从文件末尾开始读取并返回文件大小: off_t 表示文件偏移量的类型, 通常是 long / long long
    off_t file_length = lseek(fd, 0, SEEK_END);
    // 4. 给调度器分配内存
    Pager* pager = (Pager*)malloc(sizeof(Pager));
    // 5. 赋值文件描述符
    pager->file_descriptor = fd;
    // 6. 赋值文件长度
    pager->file_length = file_length;
    // 7. 初始化页内存为空
    for (uint32_t index = 0; index < TABLE_MAX_PAGES; index++) {
        pager->pages[index] = NULL;
    }
    return pager;
}

// 调度磁盘数据到页内存中
void* get_page(Pager* pager, uint32_t page_num) {
    // 1. 判断页号是否超过限制
    if (page_num > TABLE_MAX_PAGES) {
        printf("tried to fetch page number out of bounds. %d > %d\n", page_num, TABLE_MAX_PAGES);
        exit(EXIT_FAILURE);
    }
    // 2. 判断该页的数据是否已经被调度到内存中
    if (pager->pages[page_num] == NULL) {
        // 2.1 给页内存分配空间
        void* page = malloc(PAGE_SIZE);
        // 2.2 计算总的页数
        uint32_t num_pages = pager->file_length / PAGE_SIZE;
        // 2.3 判断是否新增了一页
        if (pager->file_length % PAGE_SIZE) {
            num_pages += 1;
        }
        // 2.4 判断需要加载的页内容是否为中间的页
        // TODO 页号有可能会超过总的页数吗?
        if (page_num <= num_pages) {
            // 2.4.1 从当前页开始向后读取 => 页内存时连续的
            lseek(pager->file_descriptor, page_num * PAGE_SIZE, SEEK_SET);
            // 2.4.2 开始读取页的内容
            ssize_t bytes_read = read(pager->file_descriptor, page, PAGE_SIZE);
            // 2.4.3 判断是否读取成功
            if (bytes_read == -1) {
                // TODO errno 是什么?
                printf("error reading file: %d\n", errno);
                exit(EXIT_FAILURE);
            }
        }
        pager->pages[page_num] = page;
    }
    return pager->pages[page_num];
}

// 持久化页内存的数据到磁盘中
void flush_page(Pager* pager, uint32_t page_num, uint32_t size) {
    // 1. 判断页内存指针是否为空
    if (pager->pages[page_num] == NULL) {
        printf("tried to flush null page\n");
        exit(EXIT_FAILURE);
    }
    // 2. 获取页在文件中的位置
    off_t offset = lseek(pager->file_descriptor, page_num * PAGE_SIZE, SEEK_SET);
    // 3. 判断是否读取成功
    if (offset == -1) {
        printf("error seeking: %d\n", errno);
        exit(EXIT_FAILURE);
    }
    // 4. 持久化内存
    ssize_t bytes_written = write(pager->file_descriptor, pager->pages[page_num], size);
    // 5. 判断是否写入成功
    if (bytes_written == -1) {
        printf("error writing: %d\n", errno);
        exit(EXIT_FAILURE);
    }
}

// 表结构
typedef struct {
    // 行数
    uint32_t num_rows;
    // 内存调度器
    Pager* pager;

    void *pages[TABLE_MAX_PAGES];
} Table;

// 初始化表
Table* db_open(const char* filename){
    // 1. 初始化内存调度器
    Pager* pager = pager_open(filename);
    // 2. 初始化行号
    uint32_t num_rows = pager->file_length / ROW_SIZE;
    // 3. 分配表内存
    Table* table = (Table*)malloc(sizeof(Table));
    // 4. 赋值行号
    table->num_rows = num_rows;
    // 5. 赋值内存调度器
    table->pager = pager;
    return table;
}

void db_close(Table* table) {
    // 1. 获取调度器和总页数
    Pager* pager = table->pager;
    uint32_t num_pages = table->num_rows / ROWS_PER_PAGE;
    // 2. 遍历所有页内存
    for (uint32_t index = 0; index < num_pages; index++) {
        // 2.1 如果页内存为空, 那么就跳过
        if (pager->pages[index] == NULL) {
            continue;
        }
        // 2.2 持久化页内存
        flush_page(pager, index, PAGE_SIZE);
        // 2.3 释放页内存空间
        free(pager->pages[index]);
        // 2.4 指针指向空
        pager->pages[index] = NULL;
        // 注: 先释放内存再将指针指向空, 先指针指向空会造成内存泄露
    }
    // 3. 计算是否存在跨页写入
    uint32_t num_additional_rows = table->num_rows % ROWS_PER_PAGE;
    if (num_additional_rows > 0) {
        // 3.1 跨页写入的页号就是总的页数
        uint32_t page_num = num_pages;
        // 3.2 判断需要跨页写入的数据是否还在内存中
        if (pager->pages[page_num] != NULL) {
            // 3.3 持久化需要跨页写入的数据
            flush_page(pager, page_num, num_additional_rows * ROW_SIZE);
            // 3.4 释放页内存
            free(pager->pages[page_num]);
            // 3.5 指针指向空
            pager->pages[page_num] = NULL;
        }
    }
    // 4. 关闭文件
    int result = close(pager->file_descriptor);
    // 5. 判断文件是否关闭成功
    if (result == -1) {
        printf("closing db file error.\n");
        exit(EXIT_FAILURE);
    }
    // 6. 判断是否还有没有释放的内存: 为什么还要判断呢？
    for (uint32_t index = 0; index < TABLE_MAX_PAGES; index++) {
        // 6.1 获取页内存
        void* page = pager->pages[index];
        // 6.2 判断是否已经释放过内存
        if (page) {
            free(page);
            pager->pages[index] = NULL;
        }
    }
    // 7. 释放内存调度器和表
    free(pager);
    free(table);
}


// 游标
typedef struct {
    // 表
    Table* table;
    // 行号
    uint32_t row_num;
    // 是否在表的结尾处
    bool end_of_table;
} Cursor;

// 初始化游标在表的起始位置
Cursor* table_start(Table* table) {
    // 1. 给游标分配内存
    Cursor* cursor = (Cursor*) malloc(sizeof(Cursor));
    // 2. 初始化游标属性
    cursor->table = table;
    cursor->row_num = 0;
    cursor->end_of_table = (table->num_rows == 0);
    return cursor;
}

// 初始化游标在表的结束位置
Cursor* table_end(Table* table) {
    Cursor* cursor = (Cursor*) malloc(sizeof(Cursor));
    cursor->table = table;
    cursor->row_num = table->num_rows;
    cursor->end_of_table = true;
    return cursor;
}

// 读取表中的行记录
void* cursor_value(Cursor* cursor) {
    // 1. 计算行所在的页
    uint32_t page_num = cursor->row_num / ROWS_PER_PAGE;
    // 2. 获取对应的页内存
    void* page = get_page(cursor->table->pager, page_num);
    // 3. 计算行在页中的偏移量
    uint32_t row_offset = cursor->row_num % ROWS_PER_PAGE;
    // 4. 行在页中的偏移量换算成字节偏移量
    uint32_t byte_offset = row_offset * ROW_SIZE;
    // 5. 指针运算
    return page + byte_offset;
}

// 推进游标
void cursor_advance(Cursor* cursor) {
    cursor->row_num += 1;
    if (cursor->row_num >= cursor->table->num_rows) {
        cursor->end_of_table = true;
    }
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

// SQL 执行结果
typedef enum {
    EXECUTE_SUCCESS,
    EXECUTE_TABLE_FULL
}ExecuteResult;

// 判断元命令类型
MetaCommandResult do_meta_command(InputBuffer *input_buffer, Table* table) {
    // 判断是否为退出的元命令
    if (strcmp(input_buffer->buffer, EXIT) == 0) {
        db_close(table);
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
    // 3. 游标设置到表末尾
    Cursor* cursor = table_end(table);
    // 4. 序列化行记录
    serialize_row(row_to_insert, cursor_value(cursor));
    // 5. 增加表的行记录数量
    table->num_rows += 1;
    return EXECUTE_SUCCESS;
}

// 执行 SQL 查询语句
ExecuteResult execute_select(Statement* statement, Table* table) {
    // 1. 游标设置到起始位置
    Cursor* cursor = table_start(table);
    // 2. 从游标的位置开始遍历
    Row row;
    while (!cursor->end_of_table) {
        deserialize_row(cursor_value(cursor), &row);
        print_row(&row);
        cursor_advance(cursor);
    }
    // 3. 释放游标的内存; 为什么查询需要释放内存, 插入不需要释放内存呢?
    free(cursor);
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
int main(int argc, char* argv[]) {
    // 1. 判断参数是否合法
    if (argc < 2) {
        printf("must supply a database filename.\n");
        exit(EXIT_FAILURE);
    }
    // 2. 从命令行参数中获取数据库文件名称: argv[0] 是可执行文件的名称
    char* filename = argv[1];
    // 3. 初始化表结构
    Table* table = db_open(filename);
    // 4. 初始化输入缓冲区
    InputBuffer *input_buffer = new_input_buffer();
    // 5. 持续循环读取控制台输入
    while (true) {
        // 5.1 控制台提示输入
        print_prompt();
        // 5.2 读取控制台输入
        read_input(input_buffer);
        // 5.3 判断是否为元命令 / 非 SQL 命令
        if (input_buffer->buffer[0] == '.') {
            switch (do_meta_command(input_buffer, table)) {
                case META_COMMAND_SUCCESS:
                    continue;
                case META_COMMAND_UNRECOGNIZED_COMMAND:
                    printf("unrecognized command '%s'.\n", input_buffer->buffer);
                    continue;
            }
        }
        // 5.4 初始化 SQL 语句
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
        // 5.5 执行 SQL 语句
        switch (execute_statement(&statement, table)) {
            case EXECUTE_SUCCESS:
                printf("executed.\n");
                break;
            case EXECUTE_TABLE_FULL:
                printf("error: table full. \n");
        }
    }
}
