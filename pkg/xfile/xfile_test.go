package xfile

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func getfilename() string {
	test_dir := "/tmp"
	return test_dir + "/" + "xfile_test.txt"
}

func EEq(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	} else if err1 == nil {
		return false
	} else if err2 == nil {
		return false
	}

	return err1.Error() == err2.Error()
}

func test_close(t *testing.T, xf *XFile, completed_expect bool, err_expect error) {
	completed, err := xf.Close()
	if completed != completed_expect {
		t.Errorf("completed(%v) != completed_expect(%v)", completed, completed_expect)
	}

	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	}
}

func test_block_set_uploaded(t *testing.T, xf *XFile, block_idx int, err_expect error) {
	var err = xf.SetBlockUploaded(block_idx)
	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: SetBlockUploaded ok")
	}
}

func test_block_is_uploaded(t *testing.T, xf *XFile, block_idx int, uploaded_expect bool, err_expect error) {
	var uploaded, err = xf.IsBlockUploaded(block_idx)

	if uploaded != uploaded_expect {
		t.Errorf("uploaded(%v) != uploaded_expect(%v)", uploaded, uploaded_expect)
	}
	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: IsBlockUploaded ok")
	}
}

func test_head_is_inited(t *testing.T, xf *XFile, inited_expect bool, err_expect error) {
	var inited, err = xf.HeadIsInited()
	if inited != inited_expect {
		t.Errorf("inited(%v) != inited_expect(%v)", inited, inited_expect)
	}
	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: HeadIsInited ok")
	}
}

func test_head_init(t *testing.T, xf *XFile, filesize int64, block_size int, create_time uint32, err_expect error) {
	var err = xf.HeadInit(filesize, block_size, create_time)
	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: HeadInit ok")
	}
}

//BlockWrite(block_idx int, buf []byte, size int)
func test_block_write(t *testing.T, xf *XFile, block_idx int, buf []byte, size int, err_expect error) {
	var err = xf.BlockWrite(block_idx, buf, size)
	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: BlockWrite ok")
	}
}

//BlockRead(block_idx int) ([]byte, error)
func test_block_read(t *testing.T, xf *XFile, block_idx int, buf_expect []byte, err_expect error) {
	var buf, err = xf.BlockRead(block_idx)
	sbuf := string(buf)
	sbuf_expect := string(buf_expect)
	if sbuf != sbuf_expect {
		t.Errorf("sbuf(%v) != sbuf_expect(%v)", sbuf, sbuf_expect)
	}

	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: BlockRead ok")
	}
}

func test_upload_ok(t *testing.T, xf *XFile, upload_ok_expect bool, err_expect error) {
	var upload_ok, err = xf.UploadOk()

	if upload_ok != upload_ok_expect {
		t.Errorf("upload_ok(%v) != upload_ok_expect(%v)", upload_ok, upload_ok_expect)
	}
	if !EEq(err, err_expect) {
		t.Errorf("err(%v) != err_expect(%v)", err, err_expect)
	} else {
		fmt.Println("TEST: UploadOk ok")
	}
}

func Test_OpenClose(t *testing.T) {
	var filename = getfilename()
	System("/bin/rm -f " + filename + "*")
	fmt.Println("filename: ", filename)
	var xf, err = Open(filename)
	if xf == nil {
		t.Errorf("xf is nil")
	}
	if err != nil {
		t.Errorf("err != nil, err:%v", err)
	}

	test_close(t, xf, false, nil)

	//close后进行其它操作.
	test_close(t, xf, false, fmt.Errorf("closed"))
	test_block_set_uploaded(t, xf, 1, fmt.Errorf("closed"))
	test_block_is_uploaded(t, xf, 1, false, fmt.Errorf("closed"))
	test_head_is_inited(t, xf, false, fmt.Errorf("closed"))
	test_head_init(t, xf, 0, 0, 0, fmt.Errorf("closed"))

	test_block_write(t, xf, 0, nil, 0, fmt.Errorf("closed"))
	test_block_read(t, xf, 0, nil, fmt.Errorf("closed"))
	test_upload_ok(t, xf, false, fmt.Errorf("closed"))

	// //打开失败。
	System("echo 'test' > /tmp/xfile_test.mt")
	System("echo 'test' > /tmp/xfile_test.ul")
	xf, err = Open("/tmp/xfile_test")
	if xf != nil {
		t.Errorf("xf != nil")
	}
	if !EEq(err, fmt.Errorf("no such file or directory")) {
		t.Errorf("err(%v) != '%v'", err, "no such file or directory")
	}
}

func check_head(t *testing.T, head *XFileHead, filesize int64, block_size int, create_time uint32, block_cnt int) {
	if head == nil {
		t.Error("xf.Head is nil")
	}

	if int64(head.filesize) != filesize {
		t.Errorf("head.filesize(%v) != filesize(%v)", head.filesize, filesize)
	}
	if int(head.block_size) != block_size {
		t.Errorf("head.block_size(%v) != block_size(%v)", head.block_size, block_size)
	}
	if uint32(head.create_time) != create_time {
		t.Errorf("head.create_time(%v) != create_time(%v)", head.create_time, create_time)
	}
	if int(head.block_cnt) != block_cnt {
		t.Errorf("head.block_cnt(%v) != block_cnt(%v)", head.block_cnt, block_cnt)
	}
}

func Test_HeadInit(t *testing.T) {
	var filename = getfilename()
	System("rm -f " + filename + "*")
	var xf, err = Open(filename)
	if xf == nil {
		t.Errorf("xf is nil, err: %v", err)
	}
	var filesize = int64(1024*1024*5 + 333553)
	var block_size = 1024 * 64
	var create_time = uint32(33)
	// var expires = 3600
	// var last_modified = 44
	var block_cnt = int(math.Ceil(float64(filesize) / float64(block_size)))
	fmt.Println("block_cnt:", block_cnt)
	test_head_init(t, xf, filesize, block_size, create_time, nil)
	test_head_is_inited(t, xf, true, nil)

	check_head(t, xf.Head, filesize, block_size, create_time, block_cnt)

	test_close(t, xf, false, nil)

	// open已经打开的.
	xf, err = Open(filename)
	check_head(t, xf.Head, filesize, block_size, create_time, block_cnt)
	//重复head_init
	create_time = create_time + 100

	test_head_init(t, xf, filesize, block_size, create_time, nil)
	check_head(t, xf.Head, filesize, block_size, create_time, block_cnt)
}

func RandString(str []byte, size int) []byte {
	result := make([]byte, size)

	str_len := len(str)
	for i := 0; i < size; i++ {
		var idx = rand.Intn(str_len)
		result[i] = str[idx]
	}
	return result
}

func Test_BlockDownload(t *testing.T) {
	var filename = getfilename()
	System("rm -f " + filename + "*")
	var xf, err = Open(filename)
	if xf == nil {
		t.Errorf("xf == nil, err: %v", err)
	}

	var filesize = int64(30)
	var block_size = int(8)
	var create_time = uint32(1)

	var block_cnt = int(math.Ceil(float64(filesize) / float64(block_size)))

	// test_head_init(t *testing.T, xf *XFile, filesize int64, block_size int, create_time uint32, err_expect error)
	test_head_init(t, xf, filesize, block_size, create_time, nil)

	for block_idx := 0; block_idx < block_cnt; block_idx++ {
		var block_real_size = xf.GetBlockSize(block_idx)

		var strs = RandString([]byte("0123456789abcdefghijklmnopqrstuvwxyz"), block_real_size)
		test_block_write(t, xf, block_idx, strs, block_real_size, nil)
		test_block_set_uploaded(t, xf, block_idx, nil)

		test_block_is_uploaded(t, xf, block_idx, true, nil)
		test_block_read(t, xf, block_idx, strs, nil)
	}

	test_upload_ok(t, xf, true, nil)
}
