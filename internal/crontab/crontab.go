package crontab

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// TriggeredTimeFlag theTime是一个具体的时间，返回的是当前精度支持下的词时间的表达
// 目前被crontab watcher那边用来做并发安全控制的标识
// 目前支持的是分钟级别，所以返回表达式也是到分钟
// 如果后续改到秒，此地修改就行。避免还要到对应使用者去修改
func TriggeredTimeFlag(theTime time.Time) string {
	return theTime.Format("20060102.150405")
}

type CrontabRepresent struct {
	CrontabStr string
	schedule   cron.Schedule
}

func (cr *CrontabRepresent) IsZero() bool {
	if cr == nil {
		return true
	}
	return cr.CrontabStr == ""
}

func (cr *CrontabRepresent) Equal(crB *CrontabRepresent) bool {
	if cr == nil && crB == nil {
		return true
	}
	if cr == nil || crB == nil {
		return false
	}
	return cr.CrontabStr == crB.CrontabStr
}

func (cr *CrontabRepresent) String() string {
	if cr == nil {
		return ""
	}
	return cr.CrontabStr
}

func (cr *CrontabRepresent) IsValid() bool {
	if cr.IsZero() {
		return true
	}
	return CheckValidCrontab(cr.CrontabStr)
}

func CheckValidCrontab(crontabStr string) bool {
	if crontabStr == "" {
		return true
	}
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := specParser.Parse(crontabStr)
	return err == nil
}

func BuildCrontab(iptString string) *CrontabRepresent {
	if iptString == "" {
		return nil
	}
	iptSlice := strings.Split(iptString, " ")
	if len(iptSlice) != 5 {
		return nil
	}
	specParser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	schedule, err := specParser.Parse(iptString)
	if err != nil {
		return nil
	}
	cr := &CrontabRepresent{
		CrontabStr: iptString,
		schedule:   schedule}
	return cr
}

// 分钟  小时    dayofmonth   month    dayofweek
func (cr *CrontabRepresent) RunNow() bool {
	if cr.schedule == nil { // TODO 这里是不是不对？？
		return false
	}
	// 当前此crontab配置是否应该立即运行
	m, _ := time.ParseDuration("-1m")
	nextRunTime := cr.schedule.Next(time.Now().Add(m))
	if nextRunTime.Minute() == time.Now().Minute() {
		return true
	}
	return false
}

func (cr *CrontabRepresent) UnmarshalJSON(crontabByte []byte) error {
	crontabByte = bytes.Trim(crontabByte, "\"")
	valid := CheckValidCrontab(string(crontabByte))
	if !valid {
		return errors.New("crontab not valid")
	}
	*cr = CrontabRepresent{CrontabStr: string(crontabByte)}
	return nil
}

func (cr *CrontabRepresent) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", cr.CrontabStr)
	return []byte(str), nil
}
