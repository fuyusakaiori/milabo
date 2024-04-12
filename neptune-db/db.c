#include <stdio.h>
#include <malloc/_malloc.h>
#include <stdbool.h>
#include <string.h>
#include <stdlib.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>

#define EXIT ".exit"
#define BTREE ".btree"
#define CONSTANTS ".constants"
#define INSERT "insert"
#define SELECT "select"

#define size_of_attribute(Struct, Attribute) sizeof(((Struct*)0)->Attribute);

#define COLUMN_EMAIL_SIZE 255
#define COLUMN_USERNAME_SIZE 32
#define TABLE_MAX_PAGES 100



/**
 * 元命令类型
 */
typedef enum {
    META_COMMAND_SUCCESS,
    META_COMMAND_UNRECOGNIZED_COMMAND
} MetaCommandResult;

/**
 * SQL 初始化结果类型
 */
typedef enum {
    PREPARE_SUCCESS,
    PREPARE_STRING_TOO_LONG,
    PREPARE_SYNTAX_ERROR,
    PREPARE_NEGATIVE_ID,
    PREPARE_UNRECOGNIZED_STATEMENT
} PrepareResult;

/**
 * SQL 语句类型
 */
typedef enum {
    STATEMENT_INSERT,
    STATEMENT_SELECT,
} StatementType;

/**
 * SQL 执行结果类型
 */
typedef enum {
    EXECUTE_SUCCESS,
    EXECUTE_DUPLICATE_KEY,
    EXECUTE_TABLE_FULL
}ExecuteResult;

/**
 * 树节点类型
 */
typedef enum {
    // 非叶子节点
    TREE_NODE_INTERNAL,
    // 叶子节点
    TREE_NODE_LEAF,
} TreeNodeType;

/**
 * 输入缓冲区
 */
typedef struct {
    // 字符串缓冲区
    char *buffer;
    // 缓冲区大小
    size_t buffer_length;
    // 读取字符串的长度
    ssize_t input_length;
} InputBuffer;

/**
 * 硬编码行记录
 */
typedef struct {
    // 主键
    uint32_t id;
    // 内容
    char username[COLUMN_USERNAME_SIZE + 1];
    char email[COLUMN_EMAIL_SIZE + 1];
} Row;

/**
 * SQL 语句
 */
typedef struct {
    StatementType type;
    Row row_to_insert; // 仅用于插入语句
} Statement;

/**
 * 内存调度器
 */
typedef struct {
    // 文件描述符: 用于判断文件的状态
    int file_descriptor;
    // 文件大小
    uint32_t file_length;
    // 页数量
    uint32_t num_pages;
    // 管理的内存
    void* pages[TABLE_MAX_PAGES];
} Pager;

/**
 * 表结构
 */
typedef struct {
    // 内存调度器
    Pager* pager;
    // 注: 不再记录行数而是记录根节点所在的页码
    uint32_t root_page_num;
} Table;

/**
 * 游标
 */
typedef struct {
    // 表
    Table* table;
    // 页号
    uint32_t page_num;
    // 关键字序号
    uint32_t cell_num;
    // 是否在表的结尾处
    bool end_of_table;
} Cursor;

const uint32_t ID_SIZE = size_of_attribute(Row, id);
const uint32_t USERNAME_SIZE = size_of_attribute(Row, username);
const uint32_t EMAIL_SIZE = size_of_attribute(Row, email)
const uint32_t ID_OFFSET = 0;
const uint32_t USERNAME_OFFSET = ID_OFFSET + ID_SIZE;
const uint32_t EMAIL_OFFSET = USERNAME_OFFSET + USERNAME_SIZE;
const uint32_t ROW_SIZE = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE;

// 每页大小 4KB
const uint32_t PAGE_SIZE = 4096;
// 每页的行数
const uint32_t ROWS_PER_PAGE = PAGE_SIZE / ROW_SIZE;
// 每张表的行数
const uint32_t TABLE_MAX_ROWS = ROWS_PER_PAGE * TABLE_MAX_PAGES;

/*
 * 树节点头信息的内存布局
 */
// 树节点类型字段的大小：1B
const uint32_t TREE_NODE_TYPE_SIZE = sizeof(uint8_t);
// 树节点类型字段的偏移量
const uint32_t TREE_NODE_TYPE_OFFSET = 0;
// 树节点是否为根节点字段的大小: 1B
const uint32_t IS_ROOT_SIZE = sizeof(uint8_t);
// 树节点是否为根节点字段的偏移量
const uint32_t IS_ROOT_OFFSET = TREE_NODE_TYPE_SIZE;
// 树节点的父节点指针字段的大小
const uint32_t PARENT_POINTER_SIZE = sizeof(uint32_t);
// 树节点的父节点指针字段的偏移量
const uint32_t PARENT_POINTER_OFFSET = IS_ROOT_OFFSET + IS_ROOT_SIZE;

// 树节点通用头信息的大小：为什么原本用的是 uint8_t 类型？为什么计算 uint8_t 的大小还要使用 uint32_t 接收？
const uint32_t COMMON_TREE_NODE_HEADER_SIZE = TREE_NODE_TYPE_SIZE + IS_ROOT_SIZE + PARENT_POINTER_SIZE;

/*
 * 树叶子节点头信息的内存布局
 */
// 树的叶子节点存储的关键字数量的字段的大小
const uint32_t TREE_LEAF_NODE_NUM_CELLS_SIZE = sizeof(uint32_t);
// 树的叶子节点存储的关键字数量的字段的偏移量
const uint32_t TREE_LEAF_NODE_NUM_CELLS_OFFSET = COMMON_TREE_NODE_HEADER_SIZE;
// 树的叶子节点相邻的下一个叶子节点字段的大小（单向链表）
const uint32_t TREE_LEAF_NODE_NEXT_LEAF_SIZE = sizeof(uint32_t);
// 树的叶子节点相邻的下一个叶子节点的字段的偏移量
const uint32_t TREE_LEAF_NODE_NEXT_LEAF_OFFSET = TREE_LEAF_NODE_NUM_CELLS_OFFSET + TREE_LEAF_NODE_NUM_CELLS_SIZE;

// 树的叶子节点的头信息大小
const uint32_t COMMON_TREE_LEAF_NODE_HEADER_SIZE =
        COMMON_TREE_NODE_HEADER_SIZE + TREE_LEAF_NODE_NUM_CELLS_SIZE + TREE_LEAF_NODE_NEXT_LEAF_SIZE;

/*
 * 树叶子节点存储的内容的内存布局
 */
// 树的叶子节点中每个关键字的 key 字段的大小
const uint32_t TREE_LEAF_NODE_KEY_SIZE = sizeof(uint32_t);
// 树的叶子节点中每个关键字的 key 字段的偏移量
const uint32_t TREE_LEAF_NODE_KEY_OFFSET = 0;
// 树的叶子节点中每个关键字的 value 字段的大小: 存储的内容就是行记录
const uint32_t TREE_LEAF_NODE_VALUE_SIZE = ROW_SIZE;
// 树的叶子节点中每个关键字的 value 字段的偏移量
const uint32_t TREE_LEAF_NODE_VALUE_OFFSET = TREE_LEAF_NODE_KEY_SIZE;
// 树的叶子节点中每个关键字的大小
const uint32_t TREE_LEAF_NODE_CELL_SIZE = TREE_LEAF_NODE_KEY_SIZE + TREE_LEAF_NODE_VALUE_SIZE;
// 树的叶子节点实际存储数据的大小
const uint32_t TREE_LEAF_NODE_SPACE_FOR_CELLS = PAGE_SIZE - COMMON_TREE_LEAF_NODE_HEADER_SIZE;

// 树的叶子节点最多存储的关键字数量
const uint32_t TREE_LEAF_NODE_MAX_CELLS = TREE_LEAF_NODE_SPACE_FOR_CELLS / TREE_LEAF_NODE_CELL_SIZE;

// 分裂节点时新节点的关键字数量
const uint32_t TREE_LEAF_NODE_RIGHT_SPLIT_COUNT = (TREE_LEAF_NODE_MAX_CELLS + 1) / 2;
// 分裂节点时旧节点的关键字数量
const uint32_t TREE_LEAF_NODE_LEFT_SPLIT_COUNT = (TREE_LEAF_NODE_MAX_CELLS + 1) - TREE_LEAF_NODE_RIGHT_SPLIT_COUNT;

/*
 * 树非叶子节点的头信息内存布局
 * TODO 为什么没有左子节点?
 */
// 内部节点的关键字数量的字段大小
const uint32_t TREE_INTERNAL_NODE_NUM_KEYS_SIZE = sizeof(uint32_t);
// 内部节点的关键字数量的字段的偏移量
const uint32_t TREE_INTERNAL_NODE_NUM_KEYS_OFFSET = COMMON_TREE_NODE_HEADER_SIZE;
// 内部节点的右子节点的页号字段大小
const uint32_t TREE_INTERNAL_NODE_RIGHT_CHILD_SIZE = sizeof(uint32_t);
// 内部节点的右子节点的页号字段的偏移量
const uint32_t TREE_INTERNAL_NODE_RIGHT_CHILD_OFFSET =
        TREE_INTERNAL_NODE_NUM_KEYS_OFFSET + TREE_INTERNAL_NODE_NUM_KEYS_SIZE;
// 内部节点的头信息大小
const uint32_t TREE_INTERNAL_NODE_HEADER_SIZE = COMMON_TREE_NODE_HEADER_SIZE +
        TREE_INTERNAL_NODE_NUM_KEYS_SIZE +
        TREE_INTERNAL_NODE_RIGHT_CHILD_SIZE;

/*
 * 树非叶子节点的内存布局
 */
// 内部节点的关键字 key 字段的大小
const uint32_t TREE_INTERNAL_NODE_KEY_SIZE = sizeof(uint32_t);
// 内部节点的关键字 value 字段的大小, 也就是子节点的指针
const uint32_t TREE_INTERNAL_NODE_CHILD_SIZE = sizeof(uint32_t);
// 内部节点关键字的大小
const uint32_t TREE_INTERNAL_NODE_CELL_SIZE =
        TREE_INTERNAL_NODE_KEY_SIZE + TREE_INTERNAL_NODE_CHILD_SIZE;
// 内部节点的最大关键字数量: 实际数量应该是 (4KB - 头信息) ÷ 内部节点大小
const uint32_t TREE_INTERNAL_NODE_MAX_CELLS = 3; // (PAGE_SIZE - TREE_INTERNAL_NODE_HEADER_SIZE) / TREE_INTERNAL_NODE_CELL_SIZE;


// 初始化缓冲区
InputBuffer *new_input_buffer();
// 命令行提示输入
void print_prompt();
// 输出常量内容
void print_constants();
// 可视化树结构
void print_tree(Pager* pager, uint32_t page_num, uint32_t level);
// 格式化
void indent(uint32_t level);
// 读取命令行输入内容
void read_input(InputBuffer *input_buffer);
// // 序列化每行内容
void serialize_row(Row *source, void *destination);
// 反序列化每行内容
void deserialize_row(void *source, Row *destination);
// 打印查询的内容
void print_row(Row* row);
// 获取节点类型
TreeNodeType get_node_type(void* node);
// 设置节点类型
void set_node_type(void* node, TreeNodeType type);
// 设置根节点
void set_node_root(void* node, bool is_root);
// 判断是否为根节点
bool is_node_root(void* node);
// 获取父节点
uint32_t* node_parent(void* node) {
    return node + PARENT_POINTER_OFFSET;
}
// 初始化叶子节点
void initialize_leaf_node(void* node);
// 获取叶子节点中的关键字数量
uint32_t* leaf_node_num_cells(void* node);
// 获取叶子节点中的关键字的 key
uint32_t* leaf_node_key(void* node, uint32_t cell_num);
// 获取叶子节点中的关键字的 value
void* leaf_node_value(void* node, uint32_t cell_num);
// 查询叶子节点
Cursor* leaf_node_find(Table* table, uint32_t page_num, uint32_t key);
// 查询叶子节点的下一个相邻叶子节点
uint32_t* leaf_node_next_leaf(void* node);
// 插入叶子节点
void leaf_node_insert(Cursor* cursor, uint32_t key, Row* value);
// 分裂并插入叶子节点
void leaf_node_split_and_insert(Cursor* cursor, uint32_t key, Row* value);
// 初始化内部节点
void initialize_internal_node(void* node);
// 获取内部节点的关键字数量
uint32_t* internal_node_num_keys(void* node);
// 获取内部节点最后一个关键字
// 每个内部节点的关键字数量如果是 m, 那么每个内部节点就会有 m + 1 个子节点
// 这也就意味着最后一个关键字只能单独存放到一边
uint32_t* internal_node_right_child(void* node);
// 获取内部节点的关键字
uint32_t* internal_node_cell(void* node, uint32_t cell_num);
// 获取内部节点的关键字的 key
uint32_t* internal_node_key(void* node, uint32_t key_num);
// 获取内部节点的关键字的 value - 子节点的指针
uint32_t* internal_node_child(void* node, uint32_t child_num);
// 查询内部节点
Cursor* internal_node_find(Table* table, uint32_t page_num, uint32_t key);
// 插入内部节点
void internal_node_insert(Table* table, uint32_t parent_page_num, uint32_t child_page_num);
// 查询内部节点的关键字
uint32_t internal_node_find_child(void* node, uint32_t key);
// 创建新的节点, 也就是父节点
void create_new_root(Table* table, uint32_t right_child_page_num);
// 更新内部节点的关键字
void update_internal_node_key(void* node, uint32_t old_key, uint32_t new_key);
// 获取节点中最大的关键字
uint32_t get_node_max_key(void* node);
// 初始化内存调度器
Pager* pager_open(const char* filename);
// 调度磁盘数据到页内存中
void* get_page(Pager* pager, uint32_t page_num);
// 查询未使用的页内存的页号
uint32_t get_unused_page_num(Pager* pager);
// 持久化页内存的数据到磁盘中
void pager_flush(Pager* pager, uint32_t page_num);
// 初始化表
Table* db_open(const char* filename);
// 初始化游标在表的起始位置
Cursor* table_start(Table* table);
// 移动游标的位置
Cursor* table_find(Table* table, uint32_t key);
// 读取表中的行记录
void* cursor_value(Cursor* cursor);
// 推进游标
void cursor_advance(Cursor* cursor);
// 判断元命令类型
MetaCommandResult do_meta_command(InputBuffer *input_buffer, Table* table);
// 初始化插入 SQL 语句
PrepareResult prepare_insert(InputBuffer* input_buffer, Statement* statement);
// 初始化 SQL 语句
PrepareResult prepare_statement(InputBuffer *input_buffer, Statement *statement);
// 执行 SQL 插入语句
ExecuteResult execute_insert(Statement* statement, Table* table);
// 执行 SQL 查询语句
ExecuteResult execute_select(Statement* statement, Table* table);
// 执行 SQL 语句
ExecuteResult execute_statement(Statement *statement, Table* table);



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

// 读取命令行输入内容
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

// 打印查询的内容
void print_row(Row* row) {
    printf("(%d, %s, %s)\n", row->id, row->username, row->email);
}

/*
 * 指针在和偏移量或者长度运算时其实是十六进制数和十进制数相加
 * 十六进制和十进制数是无法直接相加的, 需要转换成二进制数才可以相加
 * 这也就导致指针和偏移量运算的结果可能和预期的有所出入
 */

// 获取节点类型
TreeNodeType get_node_type(void* node) {
    uint8_t value = *((uint8_t*) node + TREE_NODE_TYPE_OFFSET);
    return (TreeNodeType) value;
}

// 设置节点类型
void set_node_type(void* node, TreeNodeType type) {
    uint8_t value = type;
    *((uint8_t*) node + TREE_NODE_TYPE_OFFSET) = value;
}

// 设置根节点
void set_node_root(void* node, bool is_root) {
    *((uint8_t *) (node + IS_ROOT_OFFSET)) = is_root;
}

// 判断是否为根节点
bool is_node_root(void* node) {
    // 1. 取出保存的是否为根节点的布尔值
    // 注: 保存的时候是 4B, 是作者觉得比较方便, 读出来的时候就转换成 1B, 因为布尔值只需要 1B
    uint8_t value = *((uint8_t*)(node + IS_ROOT_OFFSET));
    // 2. 返回布尔值
    return (bool) value;
}

// 初始化叶子节点的数量
void initialize_leaf_node(void* node) {
    // 初始化节点关键字数量
    *leaf_node_num_cells(node) = 0;
    // 设置是否为根节点
    set_node_root(node, false);
    // 设置节点类型
    set_node_type(node, TREE_NODE_LEAF);
    // 设置叶子节点的下一个相邻节点: 0 表示没有相邻节点
    *leaf_node_next_leaf(node) = 0;
}

// 获取叶子节点中的关键字数量
uint32_t* leaf_node_num_cells(void* node) {
    // 基址 + 关键字数量字段的偏移量
    return node + TREE_LEAF_NODE_NUM_CELLS_OFFSET;
}

// 获取叶子节点中的关键字
void* leaf_node_cell(void* node, uint32_t cell_num) {
    // 基址 + 叶子节点头信息大小 + 第 i 关键字 * 关键字大小
    return node + COMMON_TREE_LEAF_NODE_HEADER_SIZE + cell_num * TREE_LEAF_NODE_CELL_SIZE;
}

// 获取叶子节点中的关键字的 key
uint32_t* leaf_node_key(void* node, uint32_t cell_num) {
    // 关键字的起始位置就是 key 的位置
    return leaf_node_cell(node, cell_num);
}

// 获取叶子节点中的关键字 value
void* leaf_node_value(void* node, uint32_t cell_num) {
    return leaf_node_cell(node, cell_num) + TREE_LEAF_NODE_KEY_SIZE;
}

// 查询叶子节点
Cursor* leaf_node_find(Table* table, uint32_t page_num, uint32_t key) {
    // 1. 获取节点
    void* node = get_page(table->pager, page_num);
    // 2. 获取关键字的数量
    uint32_t num_cells = *leaf_node_num_cells(node);
    // 3. 移动游标
    Cursor* cursor = (Cursor*) malloc(sizeof(Cursor));
    cursor->table = table;
    cursor->page_num = page_num;
    // 4. 二分搜索
    uint32_t left_index = 0;
    // 注: 如果这里使用非负整数, 那么索引不能使用 len - 1, 只能使用 len, 使用 len 的问题在于会多循环几次
    uint32_t right_index = num_cells;
    while (left_index < right_index) {
        // 5. 计算中间的索引
        uint32_t mid_index = left_index + ((right_index - left_index) >> 1);
        // 6. 获取中间的关键字的 key
        uint32_t mid_value = *leaf_node_key(node, mid_index);
        // 7. 判断是否查找到了
        if (mid_value == key) {
            cursor->cell_num = mid_index;
            return cursor;
        }
        // 8. 如果没有查找到
        if (key < mid_value) {
            right_index = mid_index;
        } else {
            left_index = mid_index + 1;
        }
    }
    cursor->cell_num = left_index;
    return cursor;
}

// 查询叶子节点相邻的下一个节点
uint32_t* leaf_node_next_leaf(void* node) {
    return node + TREE_LEAF_NODE_NEXT_LEAF_OFFSET;
}

// 插入叶子节点
void leaf_node_insert(Cursor* cursor, uint32_t key, Row* value) {
    // 1. 获取叶子节点
    void* node = get_page(cursor->table->pager, cursor->page_num);
    // 2. 获取关键字数量
    uint32_t cell_nums = *leaf_node_num_cells(node);
    // 3. 判断节点的关键字数量是否超过限制, 需要分裂
    if (cell_nums >= TREE_LEAF_NODE_MAX_CELLS) {
        // 注: 叶子节点分裂
        leaf_node_split_and_insert(cursor, key, value);
        return;
    }
    // 4. 判断应该在哪里插入关键字
    if (cursor->cell_num < cell_nums) {
        // 注: 将需要插入的位置之后的数据全部向后移动, 空出来这个位置后再插入
        for (uint32_t index = cell_nums; index > cursor->cell_num; index--) {
            memcpy(leaf_node_cell(node, index),
                   leaf_node_cell(node, index - 1), TREE_LEAF_NODE_CELL_SIZE);
        }
    }
    // 5. 增加关键字数量
    *(leaf_node_num_cells(node)) += 1;
    // 6. 保存关键字的 key
    *(leaf_node_key(node, cursor->cell_num)) = key;
    // 7. 保存关键字的 value: void* 类型无法赋值, 只能拷贝
    serialize_row(value, leaf_node_value(node, cursor->cell_num));
}

// 分裂并插入叶子节点
void leaf_node_split_and_insert(Cursor* cursor, uint32_t key, Row* value) {
    // 1. 获取需要分裂的节点, 也就是当前节点
    void* old_node = get_page(cursor->table->pager, cursor->page_num);
    // 2. 获取旧节点的最大关键字
    uint32_t old_max = get_node_max_key(old_node);
    // 3. 找到还没有使用的页内存, 作为新的节点
    uint32_t new_page_num = get_unused_page_num(cursor->table->pager);
    // 4. 获取分裂后的新节点
    void* new_node = get_page(cursor->table->pager, new_page_num);
    // 5. 初始化新生成的节点
    initialize_leaf_node(new_node);
    // 6. 设置新节点的父节点
    *node_parent(new_node) = *node_parent(old_node);
    // 7. 单向链表插入节点
    *leaf_node_next_leaf(new_node) = *leaf_node_next_leaf(old_node);
    *leaf_node_next_leaf(old_node) = new_page_num;
    // 7. 将旧节点中一半的关键字都移动到新节点中
    for (int32_t index = TREE_LEAF_NODE_MAX_CELLS; index >= 0; index--) {
        // 7.1 根据分裂的阈值判断目的节点是谁
        void* destination_node = index >= TREE_LEAF_NODE_LEFT_SPLIT_COUNT ? new_node: old_node;
        // 7.2 重新定位关键字在目的节点中的位置
        uint32_t index_within_node = index % TREE_LEAF_NODE_LEFT_SPLIT_COUNT;
        // 7.3 获取关键字应该在的地址
        void* destination = leaf_node_cell(destination_node, index_within_node);
        // 7.4 如果是新插入的关键字, 那么需要序列化到节点中
        if (index == cursor->cell_num) {
            // 注1: 此前分裂的时候只保存了 value 而没有保存 key
            // 注2: 此前取的地址是整个 cell 的地址, value 是从 key 的位置开始保存的, 所以读出来的数据会有问题
            *leaf_node_key(destination_node, index_within_node) = key;
            serialize_row(value, leaf_node_value(destination_node, index_within_node));
        } else if (index > cursor->cell_num) {
            // 6.5 如果需要移动的节点比新插入的关键字大, 那么就将 index - 1 的位置拷贝过去, 因为索引是从长度开始算的
            memcpy(destination, leaf_node_cell(old_node, index - 1), TREE_LEAF_NODE_CELL_SIZE);
        } else {
            // 6.6 如果需要移动的节点比新插入的关键字小, 那么就将 index 的位置拷贝过去, 因为已经新插入了一个关键字, 所以直接使用索引是对的
            memcpy(destination, leaf_node_cell(old_node, index), TREE_LEAF_NODE_CELL_SIZE);
        }
    }
    // 8. 更新每个节点的关键字数量
    *(leaf_node_num_cells(old_node)) = TREE_LEAF_NODE_LEFT_SPLIT_COUNT;
    *(leaf_node_num_cells(new_node)) = TREE_LEAF_NODE_RIGHT_SPLIT_COUNT;
    // 9. 生成父节点并记录分裂的两个节点
    if (is_node_root(old_node)) {
        return create_new_root(cursor->table, new_page_num);
    } else {
        // 10.1 获取旧节点的父节点的页号
        uint32_t parent_page_num = *node_parent(old_node);
        // 10.2 获取旧节点现在最大的关键字
        uint32_t new_max = get_node_max_key(old_node);
        // 10.3 获取旧节点的父节点
        void* parent = get_page(cursor->table->pager, parent_page_num);
        // 10.4 更新父节点的关键字
        update_internal_node_key(parent, old_max, new_max);
        // 10.5
        internal_node_insert(cursor->table, parent_page_num, new_page_num);
    }
}

// 初始化内部节点
void initialize_internal_node(void* node) {
    // 1. 设置是否为根节点
    set_node_root(node, false);
    // 2. 设置节点类型
    set_node_type(node, TREE_NODE_INTERNAL);
    // 3. 设置节点的关键字数量
    *internal_node_num_keys(node) = 0;
}

// 获取内部节点的关键字数量
uint32_t* internal_node_num_keys(void* node) {
    // TODO 为什么 void* 可以直接隐式转换成 uint32_t*
    return node + TREE_INTERNAL_NODE_NUM_KEYS_OFFSET;
}

// 获取内部节点最后一个关键字
uint32_t* internal_node_right_child(void* node) {
    return node + TREE_INTERNAL_NODE_RIGHT_CHILD_OFFSET;
}

// 获取内部节点的关键字
uint32_t* internal_node_cell(void* node, uint32_t cell_num) {
    return node + TREE_INTERNAL_NODE_HEADER_SIZE + cell_num * TREE_INTERNAL_NODE_CELL_SIZE;
}

// 获取内部节点的关键字的 key
uint32_t* internal_node_key(void* node, uint32_t key_num) {
    // 注: 优先存放的是指针而不是 key, 为什么要这么存呢?
    return (void*)internal_node_cell(node, key_num) + TREE_INTERNAL_NODE_CHILD_SIZE;
}

// 获取内部节点的关键字的 value - 子节点的指针
uint32_t* internal_node_child(void* node, uint32_t child_num) {
    // 1. 计算现在内部节点有多少关键字
    uint32_t num_keys = *internal_node_num_keys(node);
    // 2. 判断是否访问是否超过界限
    if (child_num > num_keys) {
        printf("tried to access child_num %d > num_keys %d\n", child_num, num_keys);
        exit(EXIT_FAILURE);
    } else if (child_num == num_keys) {
        // 3. 获取最后一个关键字
        return internal_node_right_child(node);
    } else {
        // 4. 如果不是最后一个关键字, 那么直接访问就行
        return internal_node_cell(node, child_num);
    }
}

// 查询内部节点
Cursor* internal_node_find(Table* table, uint32_t page_num, uint32_t key) {
    // 1. 获取节点的地址
    void* node = get_page(table->pager, page_num);
    // 2. 获取内部节点的关键字所在的索引
    uint32_t child_index = internal_node_find_child(node, key);
    // 3. 获取关键字对应的子节点的页号
    uint32_t child_num = *internal_node_child(node, child_index);
    // 4. 获取子节点对应的地址
    void* child = get_page(table->pager, child_num);
    // 5. 判断子节点类型: 如果是内部节点, 那么就递归执行; 如果是叶子节点就会进入此前的二分查找逻辑
    switch (get_node_type(child)) {
        case TREE_NODE_LEAF:
            return leaf_node_find(table, child_num, key);
        case TREE_NODE_INTERNAL:
            return internal_node_find(table, child_num, key);
    }
}

// 插入内部节点
void internal_node_insert(Table* table, uint32_t parent_page_num, uint32_t child_page_num) {
    // 1. 获取父节点
    void* parent = get_page(table->pager, parent_page_num);
    // 2. 获取子节点
    void* child = get_page(table->pager, child_page_num);
    // 3. 获取子节点最大的关键字
    uint32_t child_max_key = get_node_max_key(child);
    // 4. 查询子节点的最大关键字在父节点中的什么位置
    uint32_t index = internal_node_find_child(parent, child_max_key);
    // 5. 获取父节点的关键字数量
    uint32_t original_num_keys = *internal_node_num_keys(parent);
    // 6. 增加父节点的关键字数量
    *internal_node_num_keys(parent) = original_num_keys + 1;
    // 7. 判断父节点的关键字数量是否超过限制
    if (original_num_keys >= TREE_INTERNAL_NODE_MAX_CELLS) {
        printf("need to implement splitting internal node\n");
        exit(EXIT_FAILURE);
    }
    // 8. 获取父节点的最右侧的子节点的页号
    uint32_t right_child_page_num = *internal_node_right_child(parent);
    // 9. 获取子节点的地址
    void* right_child = get_page(table->pager, right_child_page_num);
    // 10. 判断新加入的关键字是否比当前的最大关键字大
    if (child_max_key > get_node_max_key(right_child)) {
        // 最右侧的子节点不再是最右的了
        *internal_node_child(parent, original_num_keys) = right_child_page_num;
        // 重新设置内部节点的最右侧的关键字: 应该是以前最右侧节点的最大关键字
        *internal_node_key(parent, original_num_keys) = get_node_max_key(right_child);
        *internal_node_right_child(parent) = child_page_num;
    } else {
        // 如果新分裂节点的最大关键字 < 最右侧子节点的最大关键字 ?
        for (uint32_t cur = original_num_keys; cur > index; cur--) {
            void* destination = internal_node_cell(parent, cur);
            void* source = internal_node_cell(parent, cur - 1);
            memcpy(destination, source, TREE_INTERNAL_NODE_CELL_SIZE);
        }
        // 设置关键字对应的子节点
        *internal_node_child(parent, index) = child_page_num;
        // 设置关键字的值
        *internal_node_key(parent, index) = child_max_key;
    }
}

// 查询内部节点的关键字
uint32_t internal_node_find_child(void* node, uint32_t key) {
    // 2. 获取关键字的数量
    uint32_t num_keys = *internal_node_num_keys(node);
    // 3. 初始化二分查找左右指针
    uint32_t left_index = 0;
    uint32_t right_index = num_keys;
    // 4. 二分查找
    while (left_index < right_index) {
        // 4.1 计算中间节点
        uint32_t mid_index = left_index + ((right_index - left_index) >> 1);
        // 4.2 获取中间节点的 key
        uint32_t mid_key = *internal_node_key(node, mid_index);
        // 4.3 判断是否找到最右侧的 key
        if (mid_key >= key) {
            right_index = mid_index;
        } else {
            left_index = mid_index + 1;
        }
    }
    return left_index;
}

// 创建新的节点
void create_new_root(Table* table, uint32_t right_child_page_num) {
    // 1. 获取原本的根节点
    void* root = get_page(table->pager, table->root_page_num);
    // 2. 获取新的节点, 也就是右边的节点, 好像没有用到?
    void* right_child = get_page(table->pager, right_child_page_num);
    // 3. 为左边的节点找一块新的内存
    uint32_t left_child_page_num = get_unused_page_num(table->pager);
    // 4. 获取新的节点, 但是是左边的节点
    void* left_child = get_page(table->pager, left_child_page_num);
    // 5. 将原本根节点的内容拷贝到左边的节点
    memcpy(left_child, root, PAGE_SIZE);
    // 6. 设置左边的节点为非根节点
    set_node_root(left_child, false);
    // 7. 重新初始化根节点
    initialize_internal_node(root);
    // 8. 重新设置旧的节点为根节点: 原来初始化的时候是叶子节点
    set_node_root(root, true);
    // 9. 重新设置根节点关键字数量为 1, 因为就分裂了两个节点, 所以根节点只有一个关键字
    *internal_node_num_keys(root) = 1;
    // 10. 设置根节点的关键字的 value - 子节点的指针, 或者说页号
    *internal_node_child(root, 0) = left_child_page_num;
    // 11. 获取左子节点的最大关键字
    uint32_t left_child_max_key = get_node_max_key(left_child);
    // 12. 重新设置根节点的关键字的 key
    *internal_node_key(root, 0) = left_child_max_key;
    // 13. 最后再将最后一个子节点设置到根节点中
    *internal_node_right_child(root) = right_child_page_num;
    // 14. 设置左子结点和右子节点的父节点
    // 注: 不要直接设置为 0, 根节点不一定就是第 0 页
    *node_parent(left_child) = table->root_page_num;
    *node_parent(right_child) = table->root_page_num;
}

// 更新内部节点关键字的 key
void update_internal_node_key(void* node, uint32_t old_key, uint32_t new_key) {
    // 1. 查询旧的关键字在节点中的位置
    uint32_t old_key_index = internal_node_find_child(node, old_key);
    // 2. 更新旧的关键字的值
    *internal_node_key(node, old_key_index) = new_key;
}

// 获取节点最大的关键字
uint32_t get_node_max_key(void* node) {
    switch (get_node_type(node)) {
        case TREE_NODE_INTERNAL:
            // 内部节点的最大关键字其实也是最后一个
            return *internal_node_key(node, *internal_node_num_keys(node) - 1);
        case TREE_NODE_LEAF:
            // 叶子节点最大的关键字一定是最后一个
            return *leaf_node_key(node, *leaf_node_num_cells(node) - 1);
    }
}

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
    // 7. 初始化页的数量
    pager->num_pages = (file_length / PAGE_SIZE);
    // 8. 判断持久化页数据时是否存在异常: 正常来说每次写入的都是一页数据, 不可能除不尽
    if (file_length % PAGE_SIZE != 0) {
        printf("db file is not a whole number of pages. corrupt file. \n");
        exit(EXIT_FAILURE);
    }
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
        // 3. 更新页的数量: 如果调用的页号大于保存的页的数量, 那就需要更新页的数量
        if (page_num >= pager->num_pages) {
            pager->num_pages = page_num + 1;
        }
    }
    return pager->pages[page_num];
}

// 获取未使用的页内存的页号
uint32_t get_unused_page_num(Pager* pager) {
    // 注: 暂时直接使用最后一页就行
    return pager->num_pages;
}

// 持久化页内存的数据到磁盘中
void pager_flush(Pager* pager, uint32_t page_num) {
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
    // 4. 持久化内存: 每次固定写入一页数据
    ssize_t bytes_written = write(pager->file_descriptor, pager->pages[page_num], PAGE_SIZE);
    // 5. 判断是否写入成功
    if (bytes_written == -1) {
        printf("error writing: %d\n", errno);
        exit(EXIT_FAILURE);
    }
}

// 初始化表
Table* db_open(const char* filename) {
    // 1. 初始化内存调度器
    Pager* pager = pager_open(filename);
    // 2. 分配表内存
    Table* table = (Table*)malloc(sizeof(Table));
    // 3. 初始化页号
    table->root_page_num = 0;
    // 4. 判断页数量是否为空
    if (pager->num_pages == 0) {
        void* root_node = get_page(pager, 0);
        // 初始化第 0 页的内存
        initialize_leaf_node(root_node);
        // 因为这是第一个节点, 虽然是叶子节点, 但是同时也是根节点
        set_node_root(root_node, true);
    }
    // 5. 赋值内存调度器
    table->pager = pager;
    return table;
}

// 关闭表
void db_close(Table* table) {
    // 1. 获取调度器和总页数
    Pager* pager = table->pager;
    // 2. 遍历所有页内存
    for (uint32_t index = 0; index < pager->num_pages; index++) {
        // 2.1 如果页内存为空, 那么就跳过
        if (pager->pages[index] == NULL) {
            continue;
        }
        // 2.2 持久化页内存
        /**
         * 原逻辑: 如果没有写满一页的话, 就需要知道需要写入多少条数据, 所以后面还会单独做模运算知道剩余多少条数据
         * 现逻辑: 不管是否写满一页, 都直接写入一页大小的数据
         */
        pager_flush(pager, index);
        // 2.3 释放页内存空间
        free(pager->pages[index]);
        // 2.4 指针指向空
        pager->pages[index] = NULL;
        // 注: 先释放内存再将指针指向空, 先指针指向空会造成内存泄露
    }
    // 3. 关闭文件
    int result = close(pager->file_descriptor);
    // 4. 判断文件是否关闭成功
    if (result == -1) {
        printf("closing db file error.\n");
        exit(EXIT_FAILURE);
    }
    // 5. 判断是否还有没有释放的内存: 为什么还要判断呢？
    for (uint32_t index = 0; index < TABLE_MAX_PAGES; index++) {
        // 5.1 获取页内存
        void* page = pager->pages[index];
        // 5.2 判断是否已经释放过内存
        if (page) {
            free(page);
            pager->pages[index] = NULL;
        }
    }
    // 6. 释放内存调度器和表
    free(pager);
    free(table);
}

// 初始化游标在表的起始位置
Cursor* table_start(Table* table) {
    // 1. 查询表中最小的行记录
    Cursor* cursor = table_find(table, 0);
    // 2. 获取游标指向的页节点的地址
    void* node = get_page(table->pager, cursor->page_num);
    // 3. 判断游标是否在末尾
    cursor->end_of_table = (*leaf_node_num_cells(node) == 0);
    // 4. 返回游标
    return cursor;
}

// 移动游标的位置
Cursor* table_find(Table* table, uint32_t key) {
    // 1. 获取根节点的页号
    uint32_t root_page_num = table->root_page_num;
    // 2. 获取根节点
    void* root_node = get_page(table->pager, root_page_num);
    // 3. 搜索
    if (get_node_type(root_node) == TREE_NODE_LEAF) {
        return leaf_node_find(table, root_page_num, key);
    } else {
        return internal_node_find(table, root_page_num, key);
    }
}

// 读取表中的行记录
void* cursor_value(Cursor* cursor) {
    // 1. 获取页号
    uint32_t page_num = cursor->page_num;
    // 2. 获取对应的页内存
    void* page = get_page(cursor->table->pager, page_num);
    // 3. 获取关键字的内容
    return leaf_node_value(page, cursor->cell_num);
}

// 推进游标
void cursor_advance(Cursor* cursor) {
    // 1. 获取页号
    uint32_t page_num = cursor->page_num;
    // 2. 获取页的内容
    void* node = get_page(cursor->table->pager, page_num);
    // 3. 增加关键字的数量
    cursor->cell_num++;
    // 4. 判断游标是否已经移动到叶子节点的末尾
    if (cursor->cell_num >= (*leaf_node_num_cells(node))) {
        // 4.1 获取叶子节点相邻的下一个叶子节点的页号
        uint32_t next_page_num = *leaf_node_next_leaf(node);
        // 4.2 判断叶子节点是否存在下一个相邻叶子节点
        if (next_page_num == 0) {
            cursor->end_of_table = true;
        } else {
            cursor->page_num = next_page_num;
            cursor->cell_num = 0;
        }
    }
}

// 判断元命令类型
MetaCommandResult do_meta_command(InputBuffer *input_buffer, Table* table) {
    // 判断是否为退出的元命令
    if (strcmp(input_buffer->buffer, EXIT) == 0) {
        db_close(table);
        exit(EXIT_SUCCESS);
        // return META_COMMAND_SUCCESS;
    } else if (strcmp(input_buffer->buffer, BTREE) == 0) {
        printf("btree:\n");
        print_tree(table->pager, 0, 0);
        return META_COMMAND_SUCCESS;
    }else if (strcmp(input_buffer->buffer, CONSTANTS) == 0) {
        printf("constants: \n");
        print_constants();
        return META_COMMAND_SUCCESS;
    }
    return META_COMMAND_UNRECOGNIZED_COMMAND;
}

// 输出常量内容
void print_constants() {
    printf("ROW_SIZE: %d\n", ROW_SIZE);
    printf("COMMON_TREE_NODE_HEADER_SIZE: %d\n", COMMON_TREE_NODE_HEADER_SIZE);
    printf("COMMON_TREE_LEAF_NODE_HEADER_SIZE: %d\n", COMMON_TREE_LEAF_NODE_HEADER_SIZE);
    printf("TREE_LEAF_NODE_CELL_SIZE: %d\n", TREE_LEAF_NODE_CELL_SIZE);
    printf("TREE_LEAF_NODE_SPACE_FOR_CELLS: %d\n", TREE_LEAF_NODE_SPACE_FOR_CELLS);
    printf("TREE_LEAF_NODE_MAX_CELLS: %d\n", TREE_LEAF_NODE_MAX_CELLS);
}

// 可视化树结构
void print_tree(Pager* pager, uint32_t page_num, uint32_t level) {
    // 1. 获取节点
    void* node = get_page(pager, page_num);
    // 2. 判断节点类型
    uint32_t num_cells, child;
    switch (get_node_type(node)) {
        // 2.1 处理叶子节点
        case TREE_NODE_LEAF:
            // 2.1.1 获取叶子节点关键字数量
            num_cells = *leaf_node_num_cells(node);
            // 2.1.2 输出节点的关键字数量
            indent(level);
            printf(" - leaf (size %d)\n", num_cells);
            // 2.1.3 循环输出关键字的 key
            for (uint32_t index = 0; index < num_cells; index++) {
                indent(level + 1);
                printf("  - %d\n", *leaf_node_key(node, index));
            }
            break;
        case TREE_NODE_INTERNAL:
            // 2.2.1 获取内部节点关键字的数量
            num_cells = *internal_node_num_keys(node);
            // 2.1.2 输出节点的关键字数量
            printf(" - internal (size %d)\n", num_cells);
            // 2.1.3 循环输出关键字的 key
            for (uint32_t index = 0; index < num_cells; index++) {
                uint32_t child_num = *internal_node_child(node, index);
                print_tree(pager, child_num, level + 1);
                indent(level + 1);
                printf("  - key %d\n", *internal_node_key(node, index));
            }
            child = *internal_node_right_child(node);
            print_tree(pager, child, level + 1);
            break;
    }
}

// 格式化
void indent(uint32_t level) {
    for (uint32_t index = 0; index < level; index++) {
        printf("  ");
    }
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
    // 1. 获取页号
    void* node = get_page(table->pager, table->root_page_num);
    // 2. 获取关键字数量
    uint32_t num_cells = (*(leaf_node_num_cells(node)));
    // 3. 获取需要插入的行记录
    Row* row_to_insert = &(statement->row_to_insert);
    // 4. 获取 id
    uint32_t key_to_insert = row_to_insert->id;
    // 5. 游标移动需要插入的位置
    Cursor *cursor = table_find(table, key_to_insert);
    // 6. 判断游标的位置是否是在末尾
    if (cursor->cell_num < num_cells) {
        // 7. 获取关键字的 key
        uint32_t key_at_index = *leaf_node_key(node, cursor->cell_num);
        // 8. 如果游标的位置不在末尾, 那么就需要判断 id 是否重复
        if (key_at_index == key_to_insert) {
            return EXECUTE_DUPLICATE_KEY;
        }
    }
    // 9. 如果游标就在末尾, 那么直接插入就行, 肯定是新的
    leaf_node_insert(cursor, row_to_insert->id, row_to_insert);
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
            case (EXECUTE_DUPLICATE_KEY):
                printf("error: duplicate key. \n");
                break;
            case EXECUTE_TABLE_FULL:
                printf("error: table full. \n");
        }
    }
}
