package opt

import (
	"github.com/fBloc/bloc-server/internal/util"
)

func OptString(opts []*Opt) string {
	var resp string
	for _, opt := range opts {
		resp += opt.String()
	}
	return resp
}

func GenOptDigest(opts []*Opt) string {
	optStr := OptString(opts)
	return util.Md5Digest(optStr)
}
