package mongodb

import (
	"fmt"
	"strings"

	"github.com/fBloc/bloc-server/internal/util"
)

type confSignature = string

type MongoConfig struct {
	Addresses      []string
	Db             string
	User           string
	Password       string
	ReplicaSetName string
}

func (mC *MongoConfig) ConnectionUrl() string {
	url := "mongodb://"
	if mC.User != "" && mC.Password != "" {
		url += util.UrlEncode(mC.User) + ":" + util.UrlEncode(mC.Password) + "@"
	}

	url += strings.Join(mC.Addresses, ",")

	// TODO handle authSource param

	if mC.ReplicaSetName != "" {
		url += "&replicaSet=" + mC.ReplicaSetName
	}
	return url
}

func (mC *MongoConfig) IsNil() bool {
	if mC == nil {
		return true
	}
	return len(mC.Addresses) == 0
}

func (mC *MongoConfig) IsReplicaSet() bool {
	return len(mC.Addresses) > 0 && mC.ReplicaSetName != ""
}

func (mC MongoConfig) Equal(anotherMC MongoConfig) bool {
	if mC.Db != anotherMC.Db ||
		mC.ReplicaSetName != anotherMC.ReplicaSetName ||
		mC.User != anotherMC.User ||
		mC.Password != anotherMC.Password {
		return false
	}
	if len(mC.Addresses) != len(anotherMC.Addresses) {
		return false
	}
	hostMapAmount := make(map[string]uint, len(mC.Addresses))
	for _, v := range mC.Addresses {
		hostMapAmount[v]++
	}
	for _, v := range anotherMC.Addresses {
		hostMapAmount[v]--
	}
	for _, v := range hostMapAmount {
		if v != 0 {
			return false
		}
	}
	return true
}

func (mC MongoConfig) signature() confSignature {
	if mC.IsNil() {
		panic("nil conf cannot gen signature")
	}
	return util.Md5Digest(
		fmt.Sprintf(
			"%s_%s_%s_%s_%s",
			strings.Join(mC.Addresses, ""),
			mC.Db, mC.User, mC.Password, mC.ReplicaSetName))
}
