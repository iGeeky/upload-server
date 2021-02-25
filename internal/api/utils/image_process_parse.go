package utils

import (
	"strconv"
	"strings"
)

type CropParams struct {
	Width          int
	Height         int
	EnlargeSmaller bool
}

// ParseImageProcessParams 解析参数. resize,fw_300,fh_200/quality,q_80
func ParseImageProcessParams(imageProcess string) (quality int, crop *CropParams) {
	crop = &CropParams{}
	args := strings.Split(imageProcess, "/")
	for _, arg := range args {
		argArr := strings.Split(arg, ",")
		argName := argArr[0]
		if argName == "resize" {
			for i := 1; i < len(argArr); i++ {
				argI := argArr[i]
				if strings.HasPrefix(argI, "fw_") {
					v, _ := strconv.ParseInt(argI[3:], 10, 32)
					if v > 0 {
						crop.Width = int(v)
					}
				}
				if strings.HasPrefix(argI, "fh_") {
					v, _ := strconv.ParseInt(argI[3:], 10, 32)
					if v > 0 {
						crop.Height = int(v)
					}
				}

			}
		} else if argName == "quality" {
			if len(argArr) == 2 {
				argI := argArr[1]
				if strings.HasPrefix(argI, "q_") {
					v, _ := strconv.ParseInt(argI[2:], 10, 32)
					if v > 0 {
						quality = int(v)
					}
				}
			}
		}
	}
	return
}
