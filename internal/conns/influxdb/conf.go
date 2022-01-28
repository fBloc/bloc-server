package influxdb

import (
	"fmt"

	"github.com/fBloc/bloc-server/internal/util"
)

type confSignature = string

type InfluxDBConfig struct {
	Address      string
	UserName     string
	Password     string // need at least 8 charactors length
	Token        string
	Organization string
}

func (conf *InfluxDBConfig) Valid() bool {
	return conf.Address != "" &&
		conf.UserName != "" &&
		conf.Password != "" &&
		conf.Token != "" &&
		conf.Organization != ""
}

func (conf *InfluxDBConfig) signature() confSignature {
	if !conf.Valid() {
		panic("nil conf cannot gen signature")
	}
	return util.Md5Digest(
		fmt.Sprintf(
			"%s_%s_%s_%s_%s",
			conf.Address,
			conf.UserName, conf.Password,
			conf.Token, conf.Organization))
}
