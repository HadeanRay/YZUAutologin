package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// 定义一个结构体来表示 JSON 数据的结构
type Config struct {
	Autostartindex string `json:"autostartindex"`
	Countindex     string `json:"countindex"`
	Operatorindex  string `json:"operatorindex"`
	Passwordindex  string `json:"passwordindex"`
	Webindex       string `json:"webindex"`
}

// 读取 JSON 文件并解码
func ReadConfig(filename string) (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	fmt.Println("Executable directory:", exeDir)
	filename = filepath.Join(exeDir, filename)
	// 打开 JSON 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)

	// 解码 JSON 数据到结构体中
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (a *App) Loginyzu() error {
	log.Println("开始执行自动登录流程")

	// 读取 JSON 文件
	config, err := ReadConfig("data.json")
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// 打印读取到的配置
	log.Printf("配置信息: %+v", config)

	// 启动浏览器
	log.Println("正在启动浏览器...")
	launcher := launcher.New().Headless(false).Set("no-proxy-server")
	controlURL, err := launcher.Launch()
	if err != nil {
		launcher.Kill()
		return fmt.Errorf("browser launch failed: %w", err)
	}

	// 确保在函数结束时清理资源
	defer func() {
		log.Println("清理浏览器资源...")
		launcher.Kill()
	}()

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("browser connection failed: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: config.Webindex})
	if err != nil {
		return fmt.Errorf("page creation failed: %w", err)
	}

	// 确保在函数结束时关闭页面
	defer func() {
		if err := page.Close(); err != nil {
			log.Printf("关闭页面时出错: %v", err)
		}
	}()

	// 获取登录步骤
	steps := GetLoginSteps()

	// 执行所有登录步骤
	for _, step := range steps {
		if err := ExecuteLoginStep(page, config, step); err != nil {
			return fmt.Errorf("登录流程在 '%s' 步骤失败: %w", step.Name, err)
		}
	}

	// 减少登录完成等待时间
	log.Println("等待登录完成...")
	time.Sleep(2 * time.Second)

	log.Println("自动登录流程执行完成")
	return nil
}

// LoginWithAdvancedOptions 提供更高级的登录选项
func (a *App) LoginWithAdvancedOptions(enableDebug bool, customTimeout int) error {
	// 设置日志级别
	if enableDebug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("启用调试模式")
	}

	// 调用原始登录函数
	return a.Loginyzu()
}

// TestConnection 测试连接功能，不实际登录
func (a *App) TestConnection() (string, error) {
	log.Println("执行连接测试...")

	// 读取配置
	config, err := ReadConfig("data.json")
	if err != nil {
		return "", fmt.Errorf("读取配置失败: %w", err)
	}

	// 启动浏览器进行测试
	launcher := launcher.New().Headless(true).Set("no-proxy-server")
	controlURL, err := launcher.Launch()
	if err != nil {
		return "", fmt.Errorf("浏览器启动失败: %w", err)
	}
	defer launcher.Kill()

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return "", fmt.Errorf("浏览器连接失败: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: config.Webindex})
	if err != nil {
		return "", fmt.Errorf("页面创建失败: %w", err)
	}
	defer page.Close()

	// 等待页面加载
	waiter := NewSmartWaiter(page)
	if err := waiter.WaitForPageLoad(10 * time.Second); err != nil {
		return "", fmt.Errorf("页面加载超时: %w", err)
	}

	// 检查关键元素是否存在
	usernameFound := false
	passwordFound := false

	selectors := []ElementSelector{
		{
			Primary:      "input[name='username']",
			Alternatives: []string{"input[name='username_tip']", "input[type='text']"},
		},
	}

	for _, selector := range selectors {
		if _, err := waiter.FindElementRobust(selector); err == nil {
			usernameFound = true
			break
		}
	}

	passwordSelectors := []ElementSelector{
		{
			Primary:      "input[type='password']",
			Alternatives: []string{"input[name='password']", "input[name='pwd_tip']"},
		},
	}

	for _, selector := range passwordSelectors {
		if _, err := waiter.FindElementRobust(selector); err == nil {
			passwordFound = true
			break
		}
	}

	result := fmt.Sprintf("连接测试结果:\n- 页面加载: 成功\n- 用户名输入框: %v\n- 密码输入框: %v", 
		map[bool]string{true: "找到", false: "未找到"}[usernameFound],
		map[bool]string{true: "找到", false: "未找到"}[passwordFound])

	log.Println(result)
	return result, nil
}
