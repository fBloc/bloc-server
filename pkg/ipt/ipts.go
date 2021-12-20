package ipt

import (
	"github.com/fBloc/bloc-server/internal/util"
)

func IptString(ipts []*Ipt) string {
	var resp string
	for _, ipt := range ipts {
		resp += ipt.String()
	}
	return resp
}

func GenIptDigest(ipts []*Ipt) string {
	iptStr := IptString(ipts)
	return util.Md5Digest(iptStr)
}
