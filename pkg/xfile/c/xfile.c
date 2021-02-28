/***********************************
 * author: liuxiaojie@yunfan.com
 * date: 20160407
 ***********************************/

#include "xfile.h"
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <errno.h>
#include <assert.h>
#include <sys/mman.h>
#include <stddef.h>
#include <stdarg.h>

#define CHAR_SIZE 8
#define BITMASK(b) (1 << ((b) % CHAR_SIZE))
#define BITSLOT(b) ((b) / CHAR_SIZE)
#define BITSET(arr, b) ((arr)[BITSLOT(b)] |= BITMASK(b))
#define BITCLEAR(arr, b) ((arr)[BITSLOT(b)] &= ~BITMASK(b))
#define BITTEST(arr, b) ((arr)[BITSLOT(b)] & BITMASK(b))
#define BITNSLOTS(nb) ((nb + CHAR_SIZE - 1)/CHAR_SIZE)
#define BITARRAY(arr, size); char arr[BITNSLOTS(size)];

#ifdef __APPLE__
#define EXPORT __attribute__((visibility("default")))
#else
#define EXPORT
#endif

#define XFILE_MAGIC "XFL"
#define XFILE_HEAD_SIZE (sizeof(xfile_head_t)-8)

//根据文件大小，块大小计算块数。
#define blocks(filesize, block_size) ((filesize+block_size-1)/block_size)

typedef struct {
	char* filename;
	char* ulfilename;
	char* metafilename;
	int ulfile;
	int metafile;

	int64_t ulfilesize;
	int64_t metafilesize;
	xfile_head_t* head; //文件头信息。
}xfile_data_t;

#define IsExist(szFileName) (access(szFileName, F_OK)==0)

EXPORT int64_t GetFileSize(int file)
{
	if(file<1){
		return -1;
	}
	struct stat info;
	if(fstat(file, &info) != -1)
	{
		return (int64_t)info.st_size;
	}

	return -1;
}

EXPORT int GetFilePath(const char* fullname, char* path, int size)
{
	if(fullname == NULL || path == NULL){
		return -1;
	}

	char* pPathEnd = strrchr((char*)fullname, '/');
	if(pPathEnd == NULL){
		return -1;
	}
	pPathEnd++;

	if(pPathEnd-fullname < size){
		size = pPathEnd-fullname;
	}

	strncpy(path, fullname, size);

	return 0;
}

EXPORT int ForceMkdir(const char* path)
{
	if(path == NULL){
		return -1;
	}
	if(strlen(path)==0){
		errno = 0;
		return 0;
	}

	if(access(path, F_OK)!=0){
		int len = (int)strlen(path);
		char* cmd = (char*)alloca(len+32);
		memset(cmd,0,len+32);

		sprintf(cmd, "mkdir -p %s", path);
		int ret = system(cmd);
		return ret/256;
	}else{
		//LOG_INFO("path %s aready exists!", path);
	}
	errno = 0;
	return 0;
}

EXPORT int x_read_head(const char* metafilename, xfile_head_t* head)
{
	if(metafilename == NULL || head == NULL){
		return -1;
	}

	if(!IsExist(metafilename)){
		return -1;
	}
	int metafile = open(metafilename, O_RDONLY);
	if(metafile == -1){
		LOG_ERROR("open file [%s] failed! err:%s", metafilename, strerror(errno));
		return -1;
	}

	int ret = read(metafile, head, sizeof(xfile_head_t));
	if(ret != sizeof(xfile_head_t)){
		LOG_ERROR("read metafile head from [%s] failed! ret=%d err:%s", metafilename, ret, strerror(errno));
		ret = -1;
	}else{
		if(strcmp(head->magic, XFILE_MAGIC)!=0){
			LOG_ERROR("read metafile head from [%s] failed! magic [0x%04x] invalid!", metafilename, head->magic);
			ret = -1;
		}else{
			ret = 0;
		}
	}
	close(metafile);

	errno = 0;
	return 0;
}

EXPORT int x_block_upload_cnt(xfile_head_t* x_head){
	int x;
	int block_upload = 0;
	for(x=0;x<x_head->block_cnt;x++){
		//已经下载完的。
		if(BITTEST(x_head->bits, x)){
			block_upload++;
		}
	}
	return block_upload;
}


EXPORT xfile_t x_open(const char* filename)
{
	if(filename == NULL){
		return NULL;
	}
	const char* tmpfilename = filename;

	xfile_data_t* xfile = (xfile_data_t*)calloc(1, sizeof(xfile_data_t));
	if(xfile == NULL){
		return NULL;
	}

	int filename_len = strlen(filename);
	xfile->filename = strndup(filename, filename_len);
	if(xfile->filename == NULL){
		x_close(xfile);
		return NULL;
	}

	int tmpfilename_len = strlen(tmpfilename);
	xfile->ulfilename = (char*)calloc(1, tmpfilename_len+8);
	if(xfile->ulfilename == NULL){
		x_close(xfile);
		return NULL;
	}
	sprintf(xfile->ulfilename, "%s.ul", tmpfilename);

	xfile->metafilename = (char*)calloc(1, tmpfilename_len+8);
	if(xfile->metafilename == NULL){
		x_close(xfile);
		return NULL;
	}
	sprintf(xfile->metafilename, "%s.mt", tmpfilename);

	//TODO: metafilename exist, ulfilename not exist.
	if(IsExist(xfile->metafilename) && !IsExist(xfile->ulfilename)){
		int ret = unlink(xfile->metafilename);
		if(ret != 0){
			LOG_ERROR("unlink(%s) failed! %d err:%s", xfile->metafilename, errno, strerror(errno));
		}else{
			LOG_INFO("unlink(%s) success!", xfile->metafilename);
		}
	}

	xfile->metafile = open(xfile->metafilename, O_RDWR, S_IRUSR|S_IWUSR|S_IRGRP|S_IROTH);
	if(xfile->metafile == -1){
		//LOG_INFO("meta file [%s] failed! err:%s", xfile->metafilename, strerror(errno));
		xfile->ulfile = -1;
		errno = 0;
		return xfile;
	}else{
		xfile->ulfile = open(xfile->ulfilename, O_RDWR, S_IRUSR|S_IWUSR|S_IRGRP|S_IROTH);
		LOG_DEBUG("open localfile [%s] file=%d", xfile->ulfilename, xfile->ulfile);
		if(xfile->ulfile == -1){
			LOG_ERROR("open file [%s] failed! err:%s", xfile->ulfilename, strerror(errno));
			x_close(xfile);
			return NULL;
		}

		int ret = 0;

		int64_t meta_file_size = GetFileSize(xfile->metafile);
		LOG_INFO("metafile [%s] size:%lld", xfile->metafilename, meta_file_size);
		xfile->metafilesize = meta_file_size;

		xfile_head_t * x_head = (xfile_head_t*)mmap(NULL, meta_file_size, PROT_READ|PROT_WRITE, MAP_SHARED, xfile->metafile, 0);
		if(x_head == MAP_FAILED){
			if(errno == ENOMEM){
				LOG_ERROR("mmap failed no memory!");
			}else{
				LOG_ERROR("mmap failed! %d err:%s", errno, strerror(errno));
			}
			x_close(xfile);
			return NULL;
		}

		xfile->head = x_head;
		//校验x_head
		if(strcmp(x_head->magic, XFILE_MAGIC)!=0){
			LOG_ERROR("meta file [%s] invalid! magic error", xfile->metafilename);
			x_close(xfile);
			errno = ENOENT;
			return NULL;
		}

		//重新计算已经下载的块数。
		// int x;
		// int block_upload = 0;
		// for(x=0;x<x_head->block_cnt;x++){
		// 	//已经下载完的。
		// 	if(BITTEST(x_head->bits, x)){
		// 		block_upload++;
		// 	}
		// }

		// x_head->block_upload = block_upload;
	}

	// errno = 0;
	return xfile;
}

EXPORT int x_upload_ok(xfile_t x)
{
	errno = 0;
	if(x == NULL){
		return 0;
	}
	xfile_data_t* xfile = (xfile_data_t*)x;

	if(xfile->head == NULL){
		return 0;
	}
	int block_upload = x_block_upload_cnt(xfile->head);
	int ret = xfile->head->block_cnt> 0 && xfile->head->block_cnt == block_upload;
	// LOG_INFO("xfile->head->block_cnt(%d) xfile->head->block_upload(%d)",
	// 	xfile->head->block_cnt, block_upload);
	return ret;
}

EXPORT int x_close(xfile_t x)
{
	xfile_head_t* head = NULL;
	int ret = 0;
	xfile_data_t* xfile = (xfile_data_t*)x;

	if(xfile == NULL){
		errno = 0;
		return 0;
	}

	assert(xfile->filename != NULL);
	assert(xfile->ulfilename != NULL);
	assert(xfile->metafilename != NULL);

	int locked = 0;
	int all_uploaded = 0;

	if(xfile->head == NULL){
		if(xfile->metafile > 0){
			//fsync(xfile->metafile);
			close(xfile->metafile);
			xfile->metafile = 0;
		}
		if(xfile->ulfile > 0){
			close(xfile->ulfile);
			xfile->ulfile = 0;
		}

		if(IsExist(xfile->ulfilename)){
			ret = unlink(xfile->ulfilename);
			if(ret != 0){
				LOG_ERROR("unlink(%s) failed! %d err:%s", xfile->ulfilename, errno, strerror(errno));
			}else{
				LOG_INFO("unlink(%s) success!", xfile->ulfilename);
			}
		}
		if(IsExist(xfile->metafilename)){
			ret = unlink(xfile->metafilename);
			if(ret != 0){
				LOG_ERROR("unlink(%s) failed! %d err:%s", xfile->metafilename, errno, strerror(errno));
			}else{
				LOG_INFO("unlink(%s) success!", xfile->metafilename);
			}
		}
	}else{
		int block_upload = x_block_upload_cnt(xfile->head);
		all_uploaded = xfile->head->block_cnt == block_upload;

		if(xfile->metafile > 0){
			//fsync(xfile->metafile);
			close(xfile->metafile);
			xfile->metafile = 0;
		}
		if(xfile->ulfile > 0){
			close(xfile->ulfile);
			xfile->ulfile = 0;
		}

		if(head != NULL){
			memcpy(head, xfile->head, sizeof(xfile_head_t));
		}
		ret = munmap(xfile->head, xfile->metafilesize);
		if(ret != 0){
			LOG_ERROR("munmap failed! %d err:%s",errno, strerror(errno));
		}
		xfile->head = NULL;
	}
	if(xfile->filename != NULL){
		free(xfile->filename);
		xfile->filename = NULL;
	}

	free(xfile);

	errno = 0;
	/**
	 * 返回0 表示关闭成功，但文件未上传完成
	 * 返回1 表示关闭成功，并且文件已经上传完成。
	 */
	return all_uploaded;
}

#define CHK_IDX(block_idx, block_cnt); \
	if(block_idx >= block_cnt){ \
		LOG_ERROR("block_index(%d) >= head->block_cnt(%d)", block_idx, block_cnt);\
		return -1;\
	}

EXPORT int x_block_is_uploaded(xfile_t x, const int block_idx)
{
	errno = 0;
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return 0;
	}
	LOG_DEBUG("block_index: %d", block_idx);

	if(xfile->head == NULL){
		return 0;
	}
	xfile_head_t* head = xfile->head;
	CHK_IDX(block_idx, head->block_cnt);

	return BITTEST(head->bits, block_idx) > 0;
}

EXPORT int x_block_set_uploaded(xfile_t x, const int block_idx)
{
	errno = 0;
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return -1;
	}
	if(xfile->head == NULL){
		LOG_ERROR("xfile->head == NULL");
		return -1;
	}
	xfile_head_t* head = xfile->head;
	CHK_IDX(block_idx, head->block_cnt);

	if(BITTEST(head->bits, block_idx)==0){
		BITSET(head->bits, block_idx);
	}
	if(xfile->filename){
		LOG_DEBUG("xfile:%s, block:%d uploaded", xfile->filename, block_idx);
	}
	//立即同步，防止状态丢失，在阿里云主机上，出现过这种情况。
	fsync(xfile->metafile);

	return 0;
}

EXPORT int x_head_is_inited(xfile_t x)
{
	errno = 0;
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return 0;
	}

	return xfile->head!=NULL;
}

#ifdef __APPLE__

int fallocate(int fd, int length)
{
	fstore_t store = {F_ALLOCATECONTIG, F_PEOFPOSMODE, 0, length};
	// Try to get a continous chunk of disk space
	int ret = fcntl(fd, F_PREALLOCATE, &store);
	if(-1 == ret){
	// OK, perhaps we are too fragmented, allocate non-continuous
	store.fst_flags = F_ALLOCATEALL;
	ret = fcntl(fd, F_PREALLOCATE, &store);
	if (-1 == ret)
	  return ret;
	}
	return ftruncate(fd, length);
}
#endif

EXPORT int x_head_init(xfile_t x, const long long filesize, const int block_size, unsigned int create_time)
{
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return -1;
	}
	if(block_size<1){
		return -1;
	}
	int blocks = blocks(filesize,block_size);
	//LOG_INFO("x_head_is_inited(x): %d", x_head_is_inited(x));

	if(x_head_is_inited(x)<1){
		int filename_len = strlen(xfile->filename);
		int tmpfilename_len = strlen(xfile->ulfilename);
		filename_len = MAX(filename_len, tmpfilename_len);
		//创建目录
		int ret = 0;
		char dirname[filename_len];
		memset(dirname, 0, filename_len);
		GetFilePath(xfile->filename, dirname, filename_len);

		ret = ForceMkdir(dirname);
		if(ret != 0){
			LOG_ERROR("ForceMkdir(%s) failed! err:%s", dirname, strerror(errno));
			return -1;
		}

		memset(dirname, 0, filename_len);
		GetFilePath(xfile->ulfilename, dirname, filename_len);
		ret = ForceMkdir(dirname);
		if(ret != 0){
			LOG_ERROR("ForceMkdir(%s) failed! err:%s", dirname, strerror(errno));
			return -1;
		}

		xfile->metafile = open(xfile->metafilename, O_CREAT|O_RDWR, S_IRUSR|S_IWUSR|S_IRGRP|S_IROTH);
		if(xfile->metafile == -1){
			LOG_ERROR("open metafile [%s] failed! err:%s", xfile->metafilename, strerror(errno));
			return -1;
		}


		xfile->ulfile = open(xfile->ulfilename, O_CREAT|O_RDWR, S_IRUSR|S_IWUSR|S_IRGRP|S_IROTH);
		LOG_DEBUG("open localfile [%s] file=%d", xfile->ulfilename, xfile->ulfile);

		if(xfile->ulfile == -1){
			LOG_ERROR("open file [%s] failed! error:%s", xfile->ulfilename, strerror(errno));
			close(xfile->metafile);
			xfile->metafile = -1;
			return -1;
		}


		int bytes = (blocks+7)/8;
		int meta_file_size = XFILE_HEAD_SIZE+bytes;

		LOG_DEBUG("xfile(%s) init filesize:%lld, block_size:%d",
						xfile->filename, filesize, block_size);

		#ifdef __APPLE__
		ret = fallocate(xfile->metafile, meta_file_size);
		#else
		ret = posix_fallocate(xfile->metafile, 0, meta_file_size);
		#endif

		if(ret != 0){
			close(xfile->metafile);
			close(xfile->ulfile);
			xfile->metafile = -1;
			xfile->ulfile = -1;
			LOG_ERROR("posix_fallocate(%s, %d) failed! ret=%d, error:%s",
							xfile->metafilename, meta_file_size, ret, strerror(errno));
			return -1;
		}

		void * shm = mmap(NULL, meta_file_size, PROT_READ|PROT_WRITE, MAP_SHARED,
								xfile->metafile, 0);
		if(shm == MAP_FAILED){
			if(errno == ENOMEM){
				LOG_ERROR("mmap failed no memory!");
			}else{
				LOG_ERROR("mmap failed! %d err:%s", errno, strerror(errno));
			}
			close(xfile->metafile);
			close(xfile->ulfile);
			xfile->metafile = -1;
			xfile->ulfile = -1;
			return -1;
		}

		xfile->ulfilesize = filesize;
		xfile->metafilesize = meta_file_size;
		xfile->head = (xfile_head_t*)shm;

		//初始化head
		sprintf(xfile->head->magic, XFILE_MAGIC);
		xfile->head->filesize = filesize;
		xfile->head->block_size = block_size;
		xfile->head->block_cnt = blocks;
		xfile->head->version = XFILE_HEAD_VERSION;
	}

	xfile->head->create_time = create_time;

	// if(hash != NULL){
	// 	strncpy(xfile->head->hash, hash, sizeof(xfile->head->hash));
	// }
	//snprintf(xfile->head->rest, sizeof(xfile->head->rest)-1, "jie123108@163.com cdn system");

	errno = 0;
	return 0;
}

EXPORT int x_block_write(xfile_t x, const int block_idx, const int writed, const char* buf, const int size)
{
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return 0;
	}
	assert(xfile->head!=NULL);

	off_t offset = (off_t)block_idx * (off_t)xfile->head->block_size+writed;

	ssize_t write_len = pwrite(xfile->ulfile, buf, size, offset);
	LOG_DEBUG("block_write(file:%d, offset:%d, size:%d)", xfile->ulfile, (int)offset, size);
	if(write_len == -1){
		LOG_ERROR("pwrite(%s, size=%d, offset=%d) failed! %d err:%s",
				xfile->ulfilename, size, (int)offset, errno, strerror(errno));
	}else{
		#if 0
		int ret = fsync(xfile->ulfile);
		if(ret != 0){
			LOG_ERROR("fsync(%s) failed! %d err:%s",
				xfile->ulfilename, errno, strerror(errno));
		}
		#endif
	}
	errno = 0;
	return (int)write_len;
}

EXPORT int x_block_read(xfile_t x, const int block_idx, char* buf, const int size)
{
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return 0;
	}
	assert(xfile->head!=NULL);
	off_t offset = (off_t)block_idx * (off_t)xfile->head->block_size;
	ssize_t ret = pread(xfile->ulfile, buf, size, offset);
	LOG_DEBUG("pread(file:%d, offset:%d, size:%d)", xfile->ulfile, (int)offset, size);
	if(ret == -1){
		LOG_ERROR("pread(%s, size=%d, offset=%d) failed! %d err:%s", xfile->ulfilename, size, (int)offset, errno, strerror(errno));
	}else{
		errno = 0;
	}
	return (int)ret;
}

EXPORT xfile_head_t* x_get_file_head(xfile_t x)
{
	xfile_data_t* xfile = (xfile_data_t*)x;
	if(xfile == NULL){
		return NULL;
	}
	return xfile->head;
}


EXPORT const char* x_so_version()
{
	static char buf[64];
	memset(buf,0, sizeof(buf));
	sprintf(buf, "%s %s", __DATE__, __TIME__);
	return buf;
}

#define LOG_BUF_LEN 2048

EXPORT void ReleasePrint(const char* LEVEL, const char* funcName,
            const char* fileName, int line,  const char* format,  ...){
    va_list   args;
    va_start(args,   format);
    char vsp_buf[LOG_BUF_LEN];
    memset(vsp_buf, 0, LOG_BUF_LEN);
    int pos = 0;
    pos = snprintf(vsp_buf + pos, LOG_BUF_LEN-pos, "%s:%s[%s:%d] ",
            LEVEL, funcName, fileName, line);
    pos += vsnprintf(vsp_buf +pos ,  LOG_BUF_LEN-pos, format , args);
    if(pos < LOG_BUF_LEN-1){
        pos += snprintf(vsp_buf + pos, LOG_BUF_LEN-pos, "\n");
    }else{
        vsp_buf[LOG_BUF_LEN-2] = '\n';
        vsp_buf[LOG_BUF_LEN-1] = 0;
        pos = LOG_BUF_LEN;
    }
    fprintf(stderr, "%.*s", pos, vsp_buf);

    va_end(args);
}


#include <unistd.h>

EXPORT int print_meta_file(const char* filename, int show_detail)
{
	int ret = 0;
	if(access(filename, F_OK)!=0){
		printf("meta file [%s] not exist!\n", filename);
		return -1;
	}

	int metafile = open(filename, O_RDONLY);
	if(metafile == -1){
		printf("open meta file [%s] failed! err: %s\n", filename, strerror(errno));
		return -1;
	}

#define BUF_SIZE (1024*1024)
	char* buf = (char*)malloc(BUF_SIZE);
	memset(buf, 0, BUF_SIZE);
	do{
		int len = read(metafile, buf, BUF_SIZE);
		if(len > 0){
			xfile_head_t* head = (xfile_head_t*)buf;
			if(strcmp(head->magic, XFILE_MAGIC)!=0){
				printf("meta file [%s] invalid! magic error\n", filename);
				ret = -1;
				break;
			}
			int block_upload = x_block_upload_cnt(head);
			float uploaded_mb = (float)block_upload*head->block_size/1024/1024;
			float uploaded_percent = (float)block_upload*100/head->block_cnt;

			printf("magic: %s, filesize: %lld, uploaded: %.2fM, percent: %.2f%%\n",
					head->magic, head->filesize, uploaded_mb, uploaded_percent);

			printf("block_size: %d, block_count: %d, block_upload: %d,  \n",
					 head->block_size, head->block_cnt, block_upload);

			printf("version: %d, create_time: %u\n",
					(int)head->version, head->create_time);

			if(show_detail){
				int i,x,idx;
				int line_cnt = 64;
				for(i=0;i<head->block_cnt;i+=line_cnt){
					char buf_upload_ok[line_cnt+2];
					memset(buf_upload_ok, ' ', line_cnt);
					buf_upload_ok[line_cnt] = 0;
					for(x=0;x<line_cnt;x++){
						idx = i+x;
						if(idx < head->block_cnt){
							int upload_ok = BITTEST(head->bits, idx) > 0;
							buf_upload_ok[x] = upload_ok?'+':'-';
						}
					}
					int end = i+line_cnt;
					if(end > head->block_cnt){
						end = head->block_cnt;
					}
					printf("%4d ~ %4dm: %s\n", (int)((uint64_t)i*head->block_size/1024/1024),
						(int)(((uint64_t)end)*head->block_size/1024/1024), buf_upload_ok);
				}
			}
		}else{
			printf("read meta file [%s] failed! err: %s\n", filename, strerror(errno));
			ret = -1;
		}
	}while(0);

	free(buf);
	close(metafile);
	return ret;
}

//#define offsetof(struct_t,member) ((size_t)(char *)&((struct_t *)0)->member)

EXPORT int metaview(const char* metafile, int show_detail)
{
	printf("---    COMPILE DATA:%s %s\n", __DATE__, __TIME__);
	printf("---   XFILE_VERSION:%d, HEAD_SIZE: %d, sizeof(xfile_head_t): %d bitmap offset: %d\n",
			(int)XFILE_HEAD_VERSION, (int)XFILE_HEAD_SIZE, (int)sizeof(xfile_head_t), (int)offsetof(xfile_head_t, bits));

  	#if 0
	printf("--- offsetof(magic): %d\n", (int)offsetof(xfile_head_t, magic));
	printf("--- offsetof(version): %d\n", (int)offsetof(xfile_head_t, version));
	printf("--- offsetof(filesize): %d\n", (int)offsetof(xfile_head_t, filesize));
	printf("--- offsetof(block_size): %d\n", (int)offsetof(xfile_head_t, block_size));
	printf("--- offsetof(block_cnt): %d\n", (int)offsetof(xfile_head_t, block_cnt));
	printf("--- offsetof(create_time): %d\n", (int)offsetof(xfile_head_t, create_time));
	printf("--- offsetof(bits): %d\n", (int)offsetof(xfile_head_t, bits));
	if(argc < 2) {
		printf("Usage: %s <xfile metafile.mt> [detail]\n", argv[0]);
		exit(1);
	}
  	#endif

	return print_meta_file(metafile, show_detail);
}
