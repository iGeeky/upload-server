package utils

import (
	"bufio"
	"bytes"

	"github.com/iGeeky/open-account/pkg/baselib/log"
	// "fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"os"

	"github.com/jie123108/imaging"
)

type Size struct {
	Width, Height int
}

func _getImageSize(reader io.Reader) (*Size, error) {
	var img, _, err = image.Decode(reader)
	if err != nil {
		return nil, err
	}
	size := &Size{img.Bounds().Dx(), img.Bounds().Dy()}
	return size, nil
}

func GetImageSizeF(path string) (*Size, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return _getImageSize(file)
}

func GetImageSize(imgContent []byte) (*Size, error) {
	return _getImageSize(bytes.NewReader(imgContent))
}

/**
src size: 160x90
input size: 200x300, output: 60x90
input size: 200x100, output: 160x80
**/
func ResizeSmaller(srcWidth, srcHeight, width, height int) (newWidth int, newHeight int) {
	if width < srcWidth && height < srcHeight {
		newWidth, newHeight = width, height
		return
	}
	ratioW := float64(width) / float64(srcWidth)
	ratioH := float64(height) / float64(srcHeight)
	if ratioW >= ratioH {
		newWidth = srcWidth
		newHeight = srcWidth * height / width
	} else {
		newHeight = srcHeight
		newWidth = srcHeight * width / height
	}
	return
}

func ResizeImg(srcImg image.Image, width, height int, enlargeSmaller bool) (newImg image.Image) {
	srcBound := srcImg.Bounds()
	if width > 0 && height == 0 { // 按宽度调整图片尺寸。
		//如果不拉伸小图片，并且原图比目标尺寸小，则尺寸保持不变。
		if !enlargeSmaller && srcBound.Dx() < width {
			width = srcBound.Dx()
		}
		newImg = imaging.Resize(srcImg, width, height, imaging.Lanczos)
	} else if width == 0 && height > 0 { // 按高度调整图片尺寸。
		//如果不拉伸小图片，并且原图比目标尺寸小，则尺寸保持不变。
		if !enlargeSmaller && srcBound.Dy() < height {
			height = srcBound.Dy()
		}
		newImg = imaging.Resize(srcImg, width, height, imaging.Lanczos)
	} else { //调整尺寸并裁剪图片
		if !enlargeSmaller {
			width, height = ResizeSmaller(srcBound.Dx(), srcBound.Dy(), width, height)
		}
		newImg = imaging.Fill(srcImg, width, height, imaging.Center, imaging.Lanczos)
	}
	return
}

func ResizeImgToBytes(srcImg image.Image, filename string, width, height int, enlargeSmaller bool, quality int) (imgBytes []byte, err error) {

	newImg := ResizeImg(srcImg, width, height, enlargeSmaller)

	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	err = imaging.Encode(bufWriter, newImg, imaging.JPEG)
	if err != nil {
		log.Errorf("imaging.Encode(src: %s, width: %d, height: %d) failed! err: %v",
			filename, width, height, err)
		return
	}
	bufWriter.Flush()
	imgBytes = buf.Bytes()
	return
}

func ResizeBytesImgToBytes(srcBytes []byte, filename string, width, height int, enlargeSmaller bool, quality int) (imgBytes []byte, err error) {
	reader := bytes.NewReader(srcBytes)
	var srcImg image.Image
	srcImg, err = imaging.Decode(reader)
	if err != nil {
		log.Errorf("imaging.Decode(%s) failed! err: %v", filename, err)
		return
	}
	imgBytes, err = ResizeImgToBytes(srcImg, filename, width, height, enlargeSmaller, quality)
	return
}

func ResizeFileToFile(src string, dst string, width int, height int, enlargeSmaller bool, quality int) error {
	srcImg, err := imaging.Open(src)
	if err != nil {
		log.Errorf("Open Img File [%s] failed! err: %v", src, err)
		return err
	}

	dstImg, err := ResizeImgToBytes(srcImg, src, width, height, enlargeSmaller, quality)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, dstImg, os.ModePerm)
	if err != nil {
		log.Errorf("Save Img To File [%s] failed! err: %v", dst, err)
		return err
	}
	return nil
}
