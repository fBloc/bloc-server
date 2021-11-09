package console_debug

import (
	"log"

	infLog "github.com/fBloc/bloc-backend-go/infrastructure/log"
	"github.com/fBloc/bloc-backend-go/value_object"
)

func init() {
	var _ infLog.Logger = &ConsoleLogRepository{}
}

type msg struct {
	Level value_object.LogLevel `json:"level"`
	Data  string                `json:"data"`
}

type ConsoleLogRepository struct {
}

func New() *ConsoleLogRepository {
	resp := &ConsoleLogRepository{}
	return resp
}

func (logger *ConsoleLogRepository) SetName(name string) {
}

func (
	logger *ConsoleLogRepository,
) Infof(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func (
	logger *ConsoleLogRepository,
) Warningf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func (
	logger *ConsoleLogRepository,
) Errorf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

func (logger *ConsoleLogRepository) ForceUpload() {
}
