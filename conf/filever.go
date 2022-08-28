package conf

import (
	"time"
)

type TFilesVer struct {
	MD5        string
	LastUpdate time.Time
}

// GetFilesVer 获取或初始化文件版本信息
func GetFilesVer(k interface{}) (ver *TFilesVer) {
	v, ok := FilesVer.Load(k)
	if ok {
		ver, ok = v.(*TFilesVer)
		if ok {
			return
		}
	}
	ver = new(TFilesVer)
	FilesVer.Store(k, ver)
	return
}
