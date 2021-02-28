package xfile

/*
#cgo CFLAGS: -I c
#cgo linux LDFLAGS: ${SRCDIR}/c/libxfile.Linux.a
#cgo windows LDFLAGS: ${SRCDIR}/c/libxfile.Windows.a
#cgo darwin LDFLAGS: ${SRCDIR}/c/libxfile.Darwin.a
#include "xfile.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

/**
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
**/
type XFileHead C.xfile_head_t

type XHead struct {
	Version    int16
	Filesize   int64
	BlockSize  int32
	BlockCnt   int32
	CreateTime uint32
}

type XFile struct {
	file     C.xfile_t
	filename string
	is_open  bool
	Head     *XFileHead
}

func Open(filename string) (*XFile, error) {
	var xf XFile
	var err error
	xf.file, err = C.x_open(C.CString(filename))

	if xf.file == nil {
		return nil, err
	} else {
		xf.filename = filename
		xf.is_open = true
		xf.get_head()
	}

	return &xf, nil
}

func Metaview(metafile string, show_detail int) (int, error) {
	ret, err := C.metaview(C.CString(metafile), C.int(show_detail))
	return int(ret), err
}

/**
 * 返回false 表示关闭成功，但文件未上传完成
 * 返回true 表示关闭成功，并且文件已经上传完成。
 */
func (xf *XFile) Close() (bool, error) {
	if xf.is_open {
		ulcompleted, err := C.x_close(xf.file)
		xf.is_open = false
		return int(ulcompleted) == 1, err
	} else {
		return false, fmt.Errorf("closed")
	}
}

func (xf *XFile) DeleteFiles() {
	var ulfilename = xf.filename + ".ul"
	var mtfilename = xf.filename + ".mt"
	os.Remove(ulfilename)
	os.Remove(mtfilename)
}

func (xf *XFile) GetFilenames() (filename, ulfilename, mtfilename string) {
	filename = xf.filename
	ulfilename = xf.filename + ".ul"
	mtfilename = xf.filename + ".mt"
	return
}

type BlockStatus struct {
	CompletedChunks    []int
	NotCompletedChunks []int
}

func (xf *XFile) GetBlockStatus() (*BlockStatus, error) {
	var status *BlockStatus = nil

	var ret error = nil
	if xf.is_open {
		head, err := xf.get_head()
		if head != nil {
			block_cnt := int(head.block_cnt)
			status = &BlockStatus{}
			status.CompletedChunks = make([]int, 0)
			status.NotCompletedChunks = make([]int, 0)

			for i := 0; i < block_cnt; i++ {
				var uploaded, err = xf.IsBlockUploaded(i)
				// fmt.Println("idx:", i, ", uploaded:", uploaded, " err:", err)
				if uploaded == false && err != nil {
					return nil, err
				} else {
					if uploaded {
						status.CompletedChunks = append(status.CompletedChunks, i)
					} else {
						status.NotCompletedChunks = append(status.NotCompletedChunks, i)
					}
				}
			}
		} else {
			ret = err
		}
	} else {
		ret = fmt.Errorf("closed")
	}
	return status, ret
}

func (xf *XFile) IsBlockUploaded(block_index int) (bool, error) {
	if xf.is_open {
		downloaded, err := C.x_block_is_uploaded(xf.file, C.int(block_index))
		return downloaded == 1, err
	} else {
		downloaded, err := false, fmt.Errorf("closed")
		return downloaded, err
	}
}

func (xf *XFile) SetBlockUploaded(block_index int) error {
	if xf.is_open {
		ret, err := C.x_block_set_uploaded(xf.file, C.int(block_index))
		if ret == 0 {
			return nil
		} else {
			return err
		}
	} else {
		return fmt.Errorf("closed")
	}
}

func (xf *XFile) HeadIsInited() (bool, error) {
	if xf.is_open {
		inited, err := C.x_head_is_inited(xf.file)
		return int(inited) == 1, err
	} else {
		return false, fmt.Errorf("closed")
	}
}

func (xf *XFile) HeadInit(filesize int64, block_size int, create_time uint32) error {
	if xf.is_open {
		ret, err := C.x_head_init(xf.file, C.longlong(filesize), C.int(block_size), C.uint(create_time))
		if ret == 0 {
			xf.get_head()
			return nil
		} else {
			return err
		}
	} else {
		return fmt.Errorf("closed")
	}
}

func (xf *XFile) BlockWrite(block_idx int, buf []byte, size int) error {
	if xf.is_open {
		p := (*C.char)(unsafe.Pointer(&buf[0]))
		ret, err := C.x_block_write(xf.file, C.int(block_idx), C.int(0), p, C.int(size))
		if int(ret) == size {
			return nil
		} else {
			return err
		}
	} else {
		return fmt.Errorf("closed")
	}
}

func (xf *XFile) BlockRead(block_idx int) ([]byte, error) {
	if xf.is_open {
		head, err := xf.get_head()
		if head == nil {
			return nil, err
		}

		block_size := xf.GetBlockSize(block_idx)
		// fmt.Println("block_idx:", block_idx, ", block_size:", block_size)

		var buf = make([]byte, block_size)
		p := (*C.char)(unsafe.Pointer(&buf[0]))
		ret, err := C.x_block_read(xf.file, C.int(block_idx), p, C.int(block_size))
		if ret == C.int(block_size) {
			return buf, nil
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("closed")
	}
}

func (xf *XFile) GetHead() (*XHead, error) {
	h, err := xf.get_head()
	if err != nil {
		return nil, err
	}
	head := &XHead{}
	head.Version = int16(h.version)
	head.Filesize = int64(h.filesize)
	head.BlockSize = int32(h.block_size)
	head.BlockCnt = int32(h.block_cnt)
	head.CreateTime = uint32(h.create_time)
	return head, nil
}

func (xf *XFile) get_head() (*XFileHead, error) {
	if xf.is_open {
		head, err := C.x_get_file_head(xf.file)
		if head != nil {
			xf.Head = (*XFileHead)(head)
		}
		return (*XFileHead)(head), err
	} else {
		return nil, fmt.Errorf("closed")
	}
}

func (xf *XFile) GetBlockSize(block_idx int) int {
	if xf.is_open {
		filesize := int64(xf.Head.filesize)
		block_size := int(xf.Head.block_size)
		block_cnt := int(xf.Head.block_cnt)

		var real_block_size = block_size
		if block_idx == block_cnt-1 {
			var tmpsize = int(filesize) % block_size
			if tmpsize > 0 {
				real_block_size = tmpsize
			}
		}

		return real_block_size
	} else {
		return -1
	}
}

func (xf *XFile) UploadOk() (bool, error) {
	if xf.is_open {
		ok, err := C.x_upload_ok(xf.file)
		if err != nil {
			return false, err
		}
		return int(ok) == 1, nil
	} else {
		return false, fmt.Errorf("closed")
	}
}

func (xf *XFile) ReadAll() ([]byte, error) {
	if xf.is_open {
		head, err := xf.get_head()
		if err != err {
			return nil, err
		}
		filesize := int64(head.filesize)
		buf := make([]byte, filesize)
		p := (*C.char)(unsafe.Pointer(&buf[0]))
		ret, err := C.x_block_read(xf.file, 0, p, C.int(filesize))
		if ret == C.int(filesize) {
			return buf, nil
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("closed")
	}
}

func (xf *XFile) GetDataFile() (file *os.File, err error) {
	if xf.is_open {
		var ulfilename = xf.filename + ".ul"
		file, err = os.Open(ulfilename)
	} else {
		err = fmt.Errorf("closed")
	}
	return
}

func (xf *XFile) SoVersion() string {
	version, err := C.x_so_version()
	if err != nil {
		fmt.Println("x_so_version: %v", err)
	}
	return C.GoString(version)
}

func System(command string) (int, error) {
	ret, err := C.system(C.CString(command))
	return int(ret), err
}
