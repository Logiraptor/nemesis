package nemesis

import (
	"crypto/md5"
	"fmt"
)

// Version specifies the 'version' of a resource.
type Version string

func GetVersion(data []byte) Version {
	hash := md5.New()
	hash.Write(data)
	return Version(fmt.Sprintf("%x", hash.Sum(nil)))
}
