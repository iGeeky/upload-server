package utils

import (
	"fmt"

	coreutils "github.com/iGeeky/open-account/pkg/baselib/utils"
)

func GetRid(ver, appID, hash string) string {
	magic := "Y0o1OlO0*XT1024"
	rid := fmt.Sprintf("%s%x", ver, MMHash([]byte(appID+"-"+hash)))
	hashStr := []byte(magic + rid + magic)
	suffix := coreutils.Sha1hex(hashStr)[0:3]
	return rid + "." + suffix
}

func HashToRid(appID, hash string) string {
	return GetRid("R0", appID, hash)
}

func IdToRid(appID, id string) string {
	return GetRid("R1", appID, id)
}
