package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	filename       string
	data           map[string]string
	lastModifyTime int64
	mux            sync.RWMutex
	notifyList     []Notifyer
}

func (c *Config) parse() (m map[string]string, err error) {
	m = make(map[string]string, 1024)

	f, err := os.Open(c.filename)
	if err != nil {
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	var lineNo int
	for {
		line, errRet := reader.ReadString('\n')
		if errRet == io.EOF {
			lineParse(&lineNo, &line, &m)
			break
		}
		if errRet != nil {
			err = errRet
			return
		}
		lineParse(&lineNo, &line, &m)
	}

	return
}

func (c *Config) GetInt(key string) (value int, err error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	str, ok := c.data[key]
	if !ok {
		err = fmt.Errorf("key [%s] not found", key)
	}
	value, err = strconv.Atoi(str)

	return
}

func (c *Config) GetIntDefault(key string, defaultInt int) (value int) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	str, ok := c.data[key]
	if !ok {
		value = defaultInt
		return
	}

	value, err := strconv.Atoi(str)
	if err != nil {
		value = defaultInt
	}
	return
}

func (c *Config) GetString(key string) (value string, err error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	value, ok := c.data[key]
	if !ok {
		err = fmt.Errorf("key [%s] not found", key)
	}
	return
}

func (c *Config) GetIStringDefault(key string, defaultStr string) (value string) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	value, ok := c.data[key]
	if !ok {
		value = defaultStr
		return
	}

	return
}

func (c *Config) reload() {
	ticker := time.NewTicker(time.Second * 5)
	for _ = range ticker.C {
		func() {
			f, err := os.Open(c.filename)
			if err != nil {
				fmt.Printf("open file error: %s\n", err)
				return
			}
			defer f.Close()

			fileInfo, err := f.Stat()
			if err != nil {
				fmt.Printf("stat file error:%s\n", err)
				return
			}

			curModifyTime := fileInfo.ModTime().Unix()
			if curModifyTime > c.lastModifyTime {
				m, err := c.parse()
				if err != nil {
					fmt.Printf("parse config error:%v\n", err)
					return
				}

				c.mux.Lock()
				c.data = m
				c.mux.Unlock()

				c.lastModifyTime = curModifyTime

				for _, n := range c.notifyList {
					n.Callback(c)
				}
			}
		}()
	}
}

type Notifyer interface {
	Callback(*Config)
}

func (c *Config) AddObserver(n Notifyer) {
	c.notifyList = append(c.notifyList, n)
}

func lineParse(lineNo *int, line *string, m *map[string]string) {
	*lineNo++

	l := strings.TrimSpace(*line)
	if len(l) == 0 || l[0] == '\n' || l[0] == '#' || l[0] == ';' {
		return
	}

	itemSlice := strings.Split(l, "=")
	if len(itemSlice) == 0 {
		fmt.Printf("invalid config, line: %d", lineNo)
		return
	}

	key := strings.TrimSpace(itemSlice[0])
	if len(key) == 0 {
		fmt.Printf("invalid config, line: %d", lineNo)
		return
	}

	if len(key) == 1 {
		(*m)[key] = ""
		return
	}

	value := strings.TrimSpace(itemSlice[1])
	(*m)[key] = value

	return
}

func NewConfig(file string) (conf *Config, err error) {
	conf = &Config{
		filename: file,
		data:     make(map[string]string, 1024),
	}

	m, err := conf.parse()
	if err != nil {
		fmt.Printf("parse conf error:%v\n", err)
		return
	}

	conf.mux.Lock()
	conf.data = m
	conf.mux.Unlock()

	go conf.reload()
	return
}

type AppConfig struct {
	listenServer string
	listenPath   string
}

type AppConfigMgr struct {
	config atomic.Value
}

var appConfigMgr = &AppConfigMgr{}

func (a *AppConfigMgr) Callback(conf *Config) {
	appConfig := &AppConfig{}
	listenServer, err := conf.GetString("listenServer")
	if err != nil {
		fmt.Printf("get listenServer err: %v\n", err)
		return
	}
	appConfig.listenServer = listenServer

	listenPath, err := conf.GetString("listenPath")
	if err != nil {
		fmt.Printf("get listenPath err: %v\n", err)
		return
	}
	appConfig.listenPath = listenPath

	appConfigMgr.config.Store(appConfig)
}

func initConfig(file string) {
	conf, err := NewConfig(file)
	if err != nil {
		fmt.Printf("read config file err: %v\n", err)
		return
	}

	conf.AddObserver(appConfigMgr)

	var appConfig AppConfig
	appConfig.listenServer, err = conf.GetString("listenServer")
	if err != nil {
		fmt.Printf("get listenServer err: %v\n", err)
		return
	}
	fmt.Println("listenServer: ", appConfig.listenServer)

	appConfig.listenPath, err = conf.GetString("listenPath")
	if err != nil {
		fmt.Printf("get listenPath err: %v\n", err)
		return
	}
	fmt.Println("listenPath: ", appConfig.listenPath)

	appConfigMgr.config.Store(&appConfig)
	fmt.Println("first load success.")

}

func run() {
	for {
		appConfig := appConfigMgr.config.Load().(*AppConfig)

		fmt.Println("listenServer:", appConfig.listenServer)
		fmt.Println("listenPath:", appConfig.listenPath)
		fmt.Printf("%v\n", "--------------------")
		time.Sleep(5 * time.Second)
	}
}