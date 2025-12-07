package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// NetworkDetector 校园网检测器
type NetworkDetector struct {
	browser *rod.Browser
	timeout time.Duration
}

// NewNetworkDetector 创建新的网络检测器
func NewNetworkDetector(timeout time.Duration) (*NetworkDetector, error) {
	launcher := launcher.New().Headless(true).Set("no-proxy-server")
	controlURL, err := launcher.Launch()
	if err != nil {
		launcher.Kill()
		return nil, fmt.Errorf("浏览器启动失败: %w", err)
	}

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		launcher.Kill()
		return nil, fmt.Errorf("浏览器连接失败: %w", err)
	}

	return &NetworkDetector{
		browser: browser,
		timeout: timeout,
	}, nil
}

// Close 关闭检测器资源
func (nd *NetworkDetector) Close() {
	if nd.browser != nil {
		nd.browser.Close()
	}
}

// DetectLoginPage 检测校园网登录页面
func (nd *NetworkDetector) DetectLoginPage() (string, error) {
	log.Println("开始检测校园网登录页面...")

	// 尝试多个可能的入口点
	testURLs := []string{
		"http://www.baidu.com",           // 常用网站
		"http://www.google.com",          // 国际网站
		"http://www.yzu.edu.cn",          // 扬州大学官网
		"http://10.10.10.10",             // 常见校园网网关
		"http://1.1.1.1",                 // 常见测试地址
		"http://captive.apple.com",       // Apple 网络检测
		"http://connectivitycheck.gstatic.com", // Google 网络检测
	}

	var loginURL string
	var lastErr error

	for _, testURL := range testURLs {
		log.Printf("尝试访问: %s", testURL)
		
		url, err := nd.tryAccessURL(testURL)
		if err != nil {
			lastErr = err
			continue
		}

		// 检查是否是登录页面
		if nd.isLoginPage(url) {
			loginURL = url
			log.Printf("找到登录页面: %s", loginURL)
			break
		}
	}

	if loginURL == "" {
		return "", fmt.Errorf("未找到登录页面，最后错误: %w", lastErr)
	}

	return loginURL, nil
}

// tryAccessURL 尝试访问URL并跟踪重定向
func (nd *NetworkDetector) tryAccessURL(initialURL string) (string, error) {
	page, err := nd.browser.Page(proto.TargetCreateTarget{URL: initialURL})
	if err != nil {
		return "", fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	// 设置超时上下文
	_, cancel := context.WithTimeout(context.Background(), nd.timeout)
	defer cancel()

	// 监听网络响应
	router := page.HijackRequests()
	router.MustAdd("*", func(ctx *rod.Hijack) {
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})
	go router.Run()

	// 等待页面加载
	err = rod.Try(func() {
		page.Timeout(nd.timeout).WaitLoad()
	})
	if err != nil {
		return "", fmt.Errorf("页面加载超时: %w", err)
	}

	// 获取最终URL
	finalURL, err := page.Eval(`() => window.location.href`)
	if err != nil {
		return "", fmt.Errorf("获取页面URL失败: %w", err)
	}

	finalURLStr := finalURL.Value.String()
	log.Printf("初始URL: %s -> 最终URL: %s", initialURL, finalURLStr)
	
	return finalURLStr, nil
}

// isLoginPage 判断页面是否为登录页面
func (nd *NetworkDetector) isLoginPage(urlStr string) bool {
	// 检查URL特征
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// 常见校园网登录页面URL特征
	loginKeywords := []string{
		"login", "auth", "portal", "认证", "登录", "connect",
		"wifilogin", "web-auth", "captive-portal",
	}

	urlLower := strings.ToLower(urlStr)
	for _, keyword := range loginKeywords {
		if strings.Contains(urlLower, keyword) {
			return true
		}
	}

	// 检查常见校园网网关地址
	commonGateways := []string{
		"10.", "192.168.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
	}

	for _, gateway := range commonGateways {
		if strings.HasPrefix(u.Host, gateway) {
			return true
		}
	}

	return false
}

// TestNetworkConnectivity 测试网络连通性
func (nd *NetworkDetector) TestNetworkConnectivity() (bool, string, error) {
	log.Println("测试网络连通性...")

	// 尝试访问一个稳定的网站
	testURL := "http://www.baidu.com"
	
	page, err := nd.browser.Page(proto.TargetCreateTarget{URL: testURL})
	if err != nil {
		return false, "", fmt.Errorf("创建测试页面失败: %w", err)
	}
	defer page.Close()

	// 设置超时
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result string
	var isConnected bool

	err = rod.Try(func() {
		page.Timeout(10 * time.Second).WaitLoad()
		
		// 检查页面标题或内容
		title, _ := page.Eval(`() => document.title`)
		location, _ := page.Eval(`() => window.location.href`)
		
		finalURL := location.Value.String()
		pageTitle := title.Value.String()
		
		if strings.Contains(finalURL, "baidu.com") || strings.Contains(pageTitle, "百度") {
			isConnected = true
			result = "网络已连接，可以正常访问互联网"
		} else {
			// 检查是否被重定向到登录页面
			if nd.isLoginPage(finalURL) {
				isConnected = false
				result = fmt.Sprintf("网络需要认证，已重定向到登录页面: %s", finalURL)
			} else {
				isConnected = false
				result = fmt.Sprintf("网络状态未知，最终URL: %s", finalURL)
			}
		}
	})

	if err != nil {
		return false, "", fmt.Errorf("网络测试失败: %w", err)
	}

	return isConnected, result, nil
}

// GetNetworkStatus 获取网络状态信息
func (nd *NetworkDetector) GetNetworkStatus() (map[string]interface{}, error) {
	log.Println("获取网络状态信息...")

	status := make(map[string]interface{})
	
	// 测试连通性
	connected, connResult, err := nd.TestNetworkConnectivity()
	if err != nil {
		return nil, err
	}

	status["connected"] = connected
	status["connectivity_result"] = connResult

	// 如果未连接，尝试检测登录页面
	if !connected {
		loginURL, err := nd.DetectLoginPage()
		if err == nil {
			status["login_url"] = loginURL
			status["needs_authentication"] = true
		} else {
			status["needs_authentication"] = false
			status["detection_error"] = err.Error()
		}
	} else {
		status["needs_authentication"] = false
	}

	return status, nil
}