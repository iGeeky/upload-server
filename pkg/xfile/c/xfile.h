#ifndef __x_FILE_H__
#define __x_FILE_H__
#include <stdint.h>
#ifdef __cplusplus
extern "C"{
#endif

#define MAX(a,b) ((a)>(b)?(a):(b))

#define XFILE_HEAD_VERSION 1

typedef void* xfile_t;
typedef struct {
    char magic[4];//"XFL\0"
    int16_t version;
    int16_t int16_rest;
    int64_t filesize; //文件大小.
    int32_t block_size; //分块大小
    int32_t block_cnt; //块总数	
    uint32_t create_time; //创建时间。   
    uint32_t int32_rest;
    char rest[32+64];
    char bits[8];
}xfile_head_t;


xfile_t x_open(const char* filename);

int x_close(xfile_t xfile);


int x_block_is_uploaded(xfile_t xfile, const int block_index);
int x_block_set_uploaded(xfile_t xfile, const int block_idx);

int x_head_is_inited(xfile_t xfile);

int x_head_init(xfile_t xfile, const long long filesize, const int block_size, unsigned int create_time);

int x_block_write(xfile_t xfile, const int block_idx, const int writed, const char* buf, const int size);

int x_block_read(xfile_t xfile, const int block_idx, char* buf, const int size);

int x_upload_ok(xfile_t xfile);

xfile_head_t* x_get_file_head(xfile_t xfile);
 
const char* x_so_version();

int metaview(const char* metafile, int show_detail);

/************** DEBUG *****************/
void ReleasePrint(const char* LEVEL, const char* funcName, 
            const char* fileName, int line,  const char* format,  ...);


#ifdef DEBUG 
#define LOG_DEBUG(format, args...) \
        ReleasePrint("DEBUG", __FUNCTION__, __FILE__, __LINE__, format, ##args)
#else
#define LOG_DEBUG(format, args...) 
#endif

#define LOG_INFO(format, args...) \
        ReleasePrint(" INFO", __FUNCTION__, __FILE__, __LINE__, format, ##args)
#define LOG_WARN(format, args...) \
        ReleasePrint(" WARN", __FUNCTION__, __FILE__, __LINE__, format, ##args)
#define LOG_ERROR(format, args...) \
        ReleasePrint("ERROR", __FUNCTION__, __FILE__, __LINE__, format, ##args)


#ifdef __cplusplus
}
#endif

#endif