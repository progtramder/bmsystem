package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	go dbRoutine()
}

type chanHandler interface {
	handle()
}

//channel 的缓冲大小直接影响响应性能，可以根据情况调节缓冲大小
var dbChannel = make(chan chanHandler, 20000)

func dbRoutine() {
	for {
		handler := <-dbChannel
		handler.handle()
	}
}

type chanRegister struct {
	school string
	event  string
	token  string
	info   bminfo
}

func (self *chanRegister) handle() {
	s := getSchool(self.school)
	bmEventList := s.GetEventList()
	bmEvent := bmEventList.GetEvent(self.event)
	if bmEvent != nil {
		bmEvent.report.serialize(self.token, bmEvent.sessions[self.info.session].Desc, self.info)
	}
}

var client = &http.Client{}

func GetOpenId(code string) (openId string) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		privateData.AppId, privateData.AppSecret, code)
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	type retJson struct {
		OpenId string `json:"openid"`
	}

	rj := retJson{}
	json.Unmarshal(content, &rj)
	openId = rj.OpenId
	return
}

const (
	errSuccess = iota
	errRepeat
	errNotStarted
	errReachLimit
	errInvalidSession
)

func Reason(errCode int) string {
	switch errCode {
	case errSuccess:
		return "报名成功"
	case errRepeat:
		return "重复报名"
	case errNotStarted:
		return "报名未开始"
	case errReachLimit:
		return "已报满"
	case errInvalidSession:
		return "场次错误"
	default:
		return "未知错误"
	}
}

type Pair struct {
	key   string
	value string
}

type bminfo struct {
	session int
	form    []Pair
}

//parse json data like {"name":"Jessica","gender":"female"} into Pair array
func (self *bminfo) Load(data []byte) {
	kv := strings.TrimSuffix(strings.TrimPrefix(string(data), "{"), "}")
	pairs := strings.Split(kv, ",")
	for _, v := range pairs {
		kv := strings.Split(v, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		pair := Pair{
			strings.TrimSuffix(strings.TrimPrefix(key, `"`), `"`),
			strings.TrimSuffix(strings.TrimPrefix(value, `"`), `"`),
		}
		if pair.key == "session" {
			session, _ := strconv.ParseInt(pair.value, 10, 32)
			self.session = int(session)
		} else {
			self.form = append(self.form, pair)
		}
	}
}

func (self *bminfo) Dump() string {
	if self.form == nil {
		return "null"
	}

	data := "{"
	for _, v := range self.form {
		data += fmt.Sprintf(`"%s":"%s"`, v.key, v.value)
		data += ","
	}
	data += fmt.Sprintf(`"session":%d`, self.session)
	data += "}"
	return data
}

func (self bminfo) Equal(info bminfo) bool {
	for i, v := range info.form {
		if v.key != self.form[i].key || v.value != self.form[i].value {
			return false
		}
	}
	return true
}

type BMEvent struct {
	sync.RWMutex
	started  bool
	report   *excel
	name     string
	poster   string
	form     []Component
	sessions []Session
	bm       map[string]bminfo
}

func (self *BMEvent) put(token string, info bminfo) int {
	self.Lock()
	defer self.Unlock()
	if !self.started {
		return errNotStarted
	}
	for k, v := range self.bm {
		if k == token || v.Equal(info) {
			return errRepeat
		}
	}
	if info.session < 0 || info.session >= len(self.sessions) {
		return errInvalidSession
	}
	s := &self.sessions[info.session]
	if s.number >= s.Limit {
		return errReachLimit
	}

	s.number++
	self.bm[token] = info
	return errSuccess
}

func (self *BMEvent) has(token string) (bminfo, bool) {
	self.RLock()
	v, ok := self.bm[token]
	self.RUnlock()
	return v, ok
}

type Session struct {
	Desc    string `yaml:"description"`
	Limit   int    `yaml:"limit"`
	EndTime string `yaml:"endtime"`
	number  int
	expire  time.Time
}

func (self Session) Expired() bool {
	return time.Now().After(self.expire)
}

type Component struct {
	Type  string   `yaml:"type"  json:"type"`
	Name  string   `yaml:"name"  json:"name"`
	Value []string `yaml:"value" json:"value"`
}

type Event struct {
	Event    string      `yaml:"event"`
	Poster   string      `yaml:"poster"`
	Sessions []Session   `yaml:"sessions"`
	Form     []Component `yaml:"form"`
}

func (e Event) Compile() error {
	if e.Event == "" || e.Form == nil || e.Sessions == nil {
		return errors.New("malformed event")
	}

	for i, v := range e.Sessions {
		tm, err := parseTime(v.EndTime)
		if err != nil {
			return errors.New(fmt.Sprintf("session:%s 结束时间 %s %s", v.Desc, v.EndTime, err.Error()))
		}
		e.Sessions[i].expire = tm
	}

	return nil
}

func (self *BMEvent) Expired() bool {
	for _, v := range self.sessions {
		if !v.Expired() {
			return false
		}
	}
	return true
}

func (self *BMEvent) Init(school string, e Event) error {
	report, err := InitReport(school, e)
	if err != nil {
		return err
	}

	self.started = false
	self.name = e.Event
	self.poster = e.Poster
	self.form = e.Form
	self.sessions = e.Sessions
	self.report = report
	self.bm = map[string]bminfo{}

	return nil
}

func (self *BMEvent) Start() {
	self.Lock()
	self.started = true
	self.Unlock()
}

func (self *BMEvent) Update(e Event) error {
	//New sessions can only be appended to last one
	//Old sessions can't be deleted or be changed with sequence
	//name-change to old session is disallowed
	if len(self.sessions) > len(e.Sessions) {
		return errors.New("short sessions")
	}
	for i, v := range self.sessions {
		if v.Desc != e.Sessions[i].Desc {
			return errors.New("sessions mismatch")
		}
	}

	self.Lock()
	defer self.Unlock()

	//Only poster and limit, endtime attribute of session can be updated
	//bm info and number of sessions will be reused
	self.poster = e.Poster
	oldSessions := self.sessions
	self.sessions = e.Sessions

	match := func(desc string) int {
		for _, v := range oldSessions {
			if v.Desc == desc {
				return v.number
			}
		}
		return -1
	}

	//Copy the number attribute from old session
	for i, v := range self.sessions {
		number := match(v.Desc)
		if number != -1 {
			self.sessions[i].number = number
		}
	}

	return nil
}

type BMEventList struct {
	sync.RWMutex
	events []*BMEvent
}

func (self *BMEventList) Reset(school string) error {
	path := systembasePath + "/event.yaml"
	setting, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	eventList := struct {
		Events []Event `yaml:"events"`
	}{}

	err = yaml.Unmarshal(setting, &eventList)
	if err != nil {
		return err
	}

	for _, v := range eventList.Events {
		if err := v.Compile(); err != nil {
			return err
		}
	}

	self.Lock()
	defer self.Unlock()
	oldEvents := self.events
	self.events = make([]*BMEvent, len(eventList.Events))

	if oldEvents == nil {
		//cold reset
		for i := range self.events {
			bmEvent := &BMEvent{}
			if err := bmEvent.Init(school, eventList.Events[i]); err != nil {
				return err
			}
			self.events[i] = bmEvent
		}
	} else {
		//hot reset, we reuse the old event object if it is not expired and
		//it's name mathces that in config file
		match := func(name string) int {
			for i, v := range oldEvents {
				if v.name == name && !v.Expired() {
					return i
				}
			}
			return -1
		}

		for i, v := range eventList.Events {
			j := match(v.Event)
			if j == -1 {
				bmEvent := &BMEvent{}
				if err := bmEvent.Init(school, v); err != nil {
					//we don't touch the old event if something wrong during reset
					self.events = oldEvents
					return err
				}
				self.events[i] = bmEvent
			} else {
				//reuse the old BMEvent object and update from new event if necessary
				self.events[i] = oldEvents[j]
				if err := self.events[i].Update(v); err != nil {
					self.events = oldEvents
					return err
				}
			}
		}
	}

	return nil
}

func (self *BMEventList) GetEvent(name string) *BMEvent {
	self.RLock()
	defer self.RUnlock()
	for _, v := range self.events {
		if v.name == name {
			return v
		}
	}

	return nil
}

type school struct {
	name        string
	bmEventList *BMEventList
}

func (s *school) GetEventList() *BMEventList {
	return s.bmEventList
}

var mutexSchool sync.RWMutex
var schools = map[string]*school{}
func getSchool(name string) *school {
	if name == "" {
		return nil
	}

	mutexSchool.RLock()
	s := schools[name]
	mutexSchool.RUnlock()
	if s != nil {
		return s
	}

	//在大多数情况下程序不会执行到这里，只有极端情况下2个以上协程
	//走到这里并且只有一个会抢到写锁并且创建school对象，所以其余的协程
	//被唤醒后需要检查是否school对象是否已经被创建
	mutexSchool.Lock()
	defer mutexSchool.Unlock()
	if s = schools[name]; s != nil {
		return s
	}
	s = &school{name: name, bmEventList: &BMEventList{}}
	err := s.bmEventList.Reset(name)
	if err != nil {
		ColorRed(fmt.Sprintf("Initialize %s error: %s", name, err.Error()))
		return nil
	}
	schools[name] = s
	return s
}

func Serialize(school, event, token string, info bminfo) {
	dbChannel <- &chanRegister{
		school,
		event,
		token,
		info,
	}
}
