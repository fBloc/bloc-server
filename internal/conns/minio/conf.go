package minio

import (
	"fmt"
	"strings"

	"github.com/fBloc/bloc-server/internal/util"
)

type confSignature = string

type MinioConfig struct {
	BucketName     string
	AccessKey      string
	AccessPassword string
	Addresses      []string
}

func (mF *MinioConfig) IsNil() bool {
	if mF == nil {
		return true
	}
	return mF.BucketName == "" ||
		mF.AccessKey == "" ||
		mF.AccessPassword == "" ||
		len(mF.Addresses) == 0
}

func (mF *MinioConfig) signature() confSignature {
	if mF.IsNil() {
		panic("nil conf cannot gen signature")
	}
	return util.Md5Digest(
		fmt.Sprintf(
			"%s_%s_%s_%s",
			strings.Join(mF.Addresses, ""),
			mF.AccessKey, mF.AccessPassword, mF.BucketName))
}
