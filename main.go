package main

import (
	"github.com/google/shlex"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/spf13/pflag"
	"github.com/u2go/go-tiny-supervisor/lib/fn"
	"github.com/u2go/go-tiny-supervisor/lib/writer"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 已经处理过的程序
var processedPrograms = map[string]bool{}

// 关闭信号
var closeCh = make(chan os.Signal, 1)

// 已经启动的程序
var startedPrograms = sync.Map{}

func main() {
	// 解析命令行参数
	args := parseArgs()
	log.Println("Args:", fn.Json(args))

	// 解析配置文件
	conf := fn.Poe1(parseConfig(args.ConfigFile))
	log.Println("Config:", fn.Json(conf))

	// 切换工作目录
	log.Println("Workdir:", conf.Supervisor.Workdir)
	fn.Poe(os.Chdir(conf.Supervisor.Workdir))

	// 启动程序
	startPrograms(conf.Programs)
}

// 解析命令行参数
func parseArgs() *Args {
	ins := &Args{}
	pflag.StringVarP(&ins.ConfigFile, "config", "c", "config.json", "config file")
	pflag.Parse()
	return ins
}

// Args 命令行参数
type Args struct {
	ConfigFile string
}

// 解析配置文件
func parseConfig(file string) (*Config, error) {
	ins := &Config{}
	parser := config.NewWithOptions("config", func(opt *config.Options) {
		opt.DecoderConfig.TagName = "yaml"
		opt.ParseEnv = true
	})
	parser.WithDriver(yaml.Driver)
	if err := parser.LoadFiles(file); err != nil {
		return nil, err
	}
	return ins, parser.BindStruct("", ins)
}

// Config 配置
type Config struct {
	// 守护进程配置
	Supervisor struct {
		// 工作目录
		Workdir string `yaml:"workdir"`
	} `yaml:"supervisor"`
	// 程序
	Programs map[string]*ConfigProgram `yaml:"programs"`
}

// ConfigProgram 程序配置
type ConfigProgram struct {
	// 命令
	Command string `yaml:"command"`
	// 成功标志
	SuccessFlag map[string]string `yaml:"success_flag"`
	// 依赖
	Dependencies []string `yaml:"dependencies"`
}

// 启动所有程序
func startPrograms(programs map[string]*ConfigProgram) {
	for name := range programs {
		startProgram(name, programs)
	}

	// 等待关闭信号
	signal.Notify(closeCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-closeCh
	log.Println("Close signal:", sig)

	// 关闭所有程序
	startedPrograms.Range(func(key, value any) bool {
		name := key.(string)
		cmd := value.(*exec.Cmd)
		log.Println("Close program:", name)
		fn.Loe(cmd.Process.Signal(sig))
		return true
	})
}

// 启动程序
func startProgram(name string, programs map[string]*ConfigProgram) {
	if processedPrograms[name] {
		return
	}
	processedPrograms[name] = true

	for _, dep := range programs[name].Dependencies {
		startProgram(dep, programs)
	}

	conf := programs[name]
	if conf.SuccessFlag == nil {
		conf.SuccessFlag = map[string]string{}
	}
	command := fn.Poe1(shlex.Split(conf.Command))
	errorSleepTime := 5 * time.Second

	// 启动程序
	ok := make(chan fn.Empty, 1)
	go func() {
		for {
			cmd := exec.Command(command[0], command[1:]...)
			stdout := writer.NewWriter(name, os.Stdout, conf.SuccessFlag["stdout"])
			cmd.Stdout = stdout
			stderr := writer.NewWriter(name, os.Stderr, conf.SuccessFlag["stderr"])
			cmd.Stderr = stderr
			err := cmd.Start()
			// 启动失败，重试
			if err != nil {
				log.Println("Start program error:", name, "error:", err, ". retry after:", errorSleepTime)
				time.Sleep(errorSleepTime)
				log.Println("Retry start program:", name)
				continue
			}
			log.Println("Start program:", name, "pid:", cmd.Process.Pid)

			// 等待程序成功
			if len(conf.SuccessFlag) > 0 {
				select {
				case <-stdout.SuccessCh:
					log.Println("Program success[check by stdout]:", name)
				case <-stderr.SuccessCh:
					log.Println("Program success[check by stderr]:", name)
				}
			} else {
				log.Println("Program success[no check]:", name)
			}
			startedPrograms.Store(name, cmd)

			if len(ok) == 0 {
				ok <- fn.Empty{}
			}

			// 等待程序退出
			err = cmd.Wait()
			log.Println("Program exit:", name, "exit code:", cmd.ProcessState.ExitCode(),
				"error:", err, ". retry after:", errorSleepTime)
			time.Sleep(errorSleepTime)
			log.Println("Retry start program:", name)
		}
	}()
	<-ok
}
