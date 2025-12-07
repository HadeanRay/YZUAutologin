package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

// ElementSelector 定义多重选择器策略
type ElementSelector struct {
	Primary   string   // 主要选择器
	Alternatives []string // 备选选择器
	Attributes []string // 可能的属性名
	TextContains []string // 可能包含的文本
}

// LoginStep 定义登录步骤
type LoginStep struct {
	Name        string
	Description string
	Execute     func(*rod.Page, *Config) error
	MaxRetries  int
	Timeout     time.Duration
}

// SmartWaiter 智能等待器
type SmartWaiter struct {
	page *rod.Page
}

// NewSmartWaiter 创建智能等待器
func NewSmartWaiter(page *rod.Page) *SmartWaiter {
	return &SmartWaiter{page: page}
}

// WaitForElementWithRetry 等待元素出现，支持重试
func (sw *SmartWaiter) WaitForElementWithRetry(selectors []string, timeout time.Duration) (*rod.Element, error) {
	var lastErr error
	
	for _, selector := range selectors {
		element, err := sw.page.Element(selector)
		if err == nil {
			return element, nil
		}
		lastErr = err
	}
	
	return nil, fmt.Errorf("none of the selectors worked: %v", lastErr)
}

// WaitForPageLoad 智能等待页面加载完成
func (sw *SmartWaiter) WaitForPageLoad(timeout time.Duration) error {
	// 等待页面完全加载
	err := sw.page.WaitLoad()
	if err != nil {
		return fmt.Errorf("page load timeout: %w", err)
	}
	
	// 额外等待一段时间确保动态内容加载完成
	time.Sleep(1 * time.Second)
	return nil
}

// FindElementRobust 健壮的元素查找
func (sw *SmartWaiter) FindElementRobust(selector ElementSelector) (*rod.Element, error) {
	// 尝试主要选择器
	if selector.Primary != "" {
		element, err := sw.page.Element(selector.Primary)
		if err == nil {
			return element, nil
		}
	}
	
	// 尝试备选选择器
	for _, alt := range selector.Alternatives {
		element, err := sw.page.Element(alt)
		if err == nil {
			return element, nil
		}
	}
	
	// 尝试通过属性查找
	for _, attr := range selector.Attributes {
		elements, err := sw.page.Elements(fmt.Sprintf("[%s]", attr))
		if err == nil && len(elements) > 0 {
			return elements[0], nil
		}
	}
	
	// 尝试通过文本内容查找
	for _, text := range selector.TextContains {
		element, err := sw.page.Element(fmt.Sprintf("//*[contains(text(), '%s')]", text))
		if err == nil {
			return element, nil
		}
	}
	
	return nil, fmt.Errorf("element not found with any selector")
}

// RetryOperation 带重试的操作
func RetryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		if i < maxRetries-1 {
			log.Printf("操作失败，%v 后重试 (尝试 %d/%d): %v", delay, i+1, maxRetries, err)
			time.Sleep(delay)
		}
	}
	
	return fmt.Errorf("操作在 %d 次尝试后仍然失败: %w", maxRetries, lastErr)
}

// ExecuteLoginStep 执行登录步骤
func ExecuteLoginStep(page *rod.Page, config *Config, step LoginStep) error {
	log.Printf("执行步骤: %s - %s", step.Name, step.Description)
	
	operation := func() error {
		return step.Execute(page, config)
	}
	
	err := RetryOperation(operation, step.MaxRetries, step.Timeout/4)
	if err != nil {
		return fmt.Errorf("步骤 '%s' 失败: %w", step.Name, err)
	}
	
	log.Printf("步骤 '%s' 成功完成", step.Name)
	return nil
}

// GetLoginSteps 获取登录步骤序列
func GetLoginSteps() []LoginStep {
	return []LoginStep{
		{
			Name:        "等待页面加载",
			Description: "等待登录页面完全加载",
			Execute:     waitForPageLoad,
			MaxRetries:  3,
			Timeout:     10 * time.Second,
		},
		{
			Name:        "输入用户名",
			Description: "在用户名输入框中输入账号",
			Execute:     inputUsername,
			MaxRetries:  3,
			Timeout:     5 * time.Second,
		},
		{
			Name:        "输入密码",
			Description: "在密码输入框中输入密码",
			Execute:     inputPassword,
			MaxRetries:  3,
			Timeout:     5 * time.Second,
		},
		{
			Name:        "提交表单",
			Description: "提交登录表单",
			Execute:     submitForm,
			MaxRetries:  2,
			Timeout:     5 * time.Second,
		},
		{
			Name:        "选择运营商",
			Description: "选择网络运营商",
			Execute:     selectOperator,
			MaxRetries:  3,
			Timeout:     5 * time.Second,
		},
		{
			Name:        "确认登录",
			Description: "点击最终登录按钮",
			Execute:     confirmLogin,
			MaxRetries:  3,
			Timeout:     5 * time.Second,
		},
	}
}

// waitForPageLoad 等待页面加载
func waitForPageLoad(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	return waiter.WaitForPageLoad(10 * time.Second)
}

// inputUsername 输入用户名
func inputUsername(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	
	// 定义多种可能的选择器
	selectors := []ElementSelector{
		{
			Primary:      "input[name='username']",
			Alternatives: []string{"input[name='username_tip']", "input[type='text']", "input[id*='username']", "input[class*='username']"},
			Attributes:   []string{"placeholder", "name", "id"},
			TextContains: []string{"用户名", "账号", "学号", "username"},
		},
	}
	
	for _, selector := range selectors {
		element, err := waiter.FindElementRobust(selector)
		if err == nil {
			// 清空输入框并输入新内容
			if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
				return fmt.Errorf("点击用户名输入框失败: %w", err)
			}
			
			// 全选并清空
			if err := element.SelectAllText(); err != nil {
				return fmt.Errorf("选择文本失败: %w", err)
			}
			
			if err := element.Input(""); err != nil {
				return fmt.Errorf("清空输入框失败: %w", err)
			}
			
			// 输入用户名
			if err := element.Input(config.Countindex); err != nil {
				return fmt.Errorf("输入用户名失败: %w", err)
			}
			
			log.Printf("成功输入用户名")
			return nil
		}
	}
	
	return fmt.Errorf("找不到用户名输入框")
}

// inputPassword 输入密码
func inputPassword(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	
	// 定义多种可能的选择器
	selectors := []ElementSelector{
		{
			Primary:      "input[type='password']",
			Alternatives: []string{"input[name='password']", "input[name='pwd_tip']", "input[type='password']", "input[id*='password']", "input[class*='password']"},
			Attributes:   []string{"type", "name", "id"},
			TextContains: []string{"密码", "password"},
		},
	}
	
	for _, selector := range selectors {
		element, err := waiter.FindElementRobust(selector)
		if err == nil {
			// 点击输入框
			if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
				return fmt.Errorf("点击密码输入框失败: %w", err)
			}
			
			// 全选并清空
			if err := element.SelectAllText(); err != nil {
				return fmt.Errorf("选择密码文本失败: %w", err)
			}
			
			if err := element.Input(""); err != nil {
				return fmt.Errorf("清空密码输入框失败: %w", err)
			}
			
			// 输入密码
			if err := element.Input(config.Passwordindex); err != nil {
				return fmt.Errorf("输入密码失败: %w", err)
			}
			
			log.Printf("成功输入密码")
			return nil
		}
	}
	
	return fmt.Errorf("找不到密码输入框")
}

// submitForm 提交表单
func submitForm(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	
	// 尝试多种提交方式
	submitWays := []func() error{
		// 方式1: 按回车
		func() error {
			element, err := waiter.FindElementRobust(ElementSelector{
				Primary:      "input[type='password']",
				Alternatives: []string{"input[name='password']", "input[name='pwd_tip']"},
			})
			if err != nil {
				return err
			}
			return element.Type(input.Enter)
		},
		// 方式2: 点击登录按钮
		func() error {
			selectors := []string{
				"input[type='submit']",
				"button[type='submit']",
				"input[value*='登录']",
				"button:contains('登录')",
				"input[value*='Login']",
				"button:contains('Login')",
				"#logincompus_pc_hk_1",
			}
			
			for _, selector := range selectors {
				element, err := page.Element(selector)
				if err == nil {
					return element.Click(proto.InputMouseButtonLeft, 1)
				}
			}
			return fmt.Errorf("找不到登录按钮")
		},
	}
	
	for _, submitFunc := range submitWays {
		if err := submitFunc(); err == nil {
			log.Printf("成功提交表单")
			return nil
		}
	}
	
	return fmt.Errorf("无法提交表单")
}

// selectOperator 选择运营商
func selectOperator(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	
	// 等待页面稳定
	time.Sleep(2 * time.Second)
	
	// 点击下拉菜单
	selectors := []ElementSelector{
		{
			Primary:      "#selectDisname",
			Alternatives: []string{"select[name='operator']", "select[id*='operator']", ".select-operator"},
			Attributes:   []string{"id", "name", "class"},
			TextContains: []string{"选择", "运营商", "请选择"},
		},
	}
	
	for _, selector := range selectors {
		element, err := waiter.FindElementRobust(selector)
		if err == nil {
			if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
				continue
			}
			log.Printf("成功点击运营商选择框")
			break
		}
	}
	
	// 等待下拉选项加载
	time.Sleep(1 * time.Second)
	
	// 选择具体的运营商
	var serviceSelector string
	switch config.Operatorindex {
	case "a":
		serviceSelector = "#_service_0"
	case "b":
		serviceSelector = "#_service_1"
	case "c":
		serviceSelector = "#_service_2"
	case "d":
		serviceSelector = "#_service_3"
	default:
		return fmt.Errorf("无效的运营商索引: %s", config.Operatorindex)
	}
	
	// 尝试多种选择器
	operatorSelectors := []string{
		serviceSelector,
		fmt.Sprintf("input[value='%s']", config.Operatorindex),
		fmt.Sprintf("option[value='%s']", config.Operatorindex),
	}
	
	for _, selector := range operatorSelectors {
		element, err := page.Element(selector)
		if err == nil {
			if err := element.Click(proto.InputMouseButtonLeft, 1); err == nil {
				log.Printf("成功选择运营商")
				return nil
			}
		}
	}
	
	return fmt.Errorf("无法选择运营商")
}

// confirmLogin 确认登录
func confirmLogin(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	
	// 等待页面稳定
	time.Sleep(1 * time.Second)
	
	// 查找最终登录按钮
	selectors := []ElementSelector{
		{
			Primary:      "#loginLink",
			Alternatives: []string{"input[type='submit']", "button[type='submit']", "button:contains('登录')", "a:contains('登录')", ".login-btn"},
			Attributes:   []string{"id", "class", "type"},
			TextContains: []string{"登录", "Login", "连接", "确认"},
		},
	}
	
	for _, selector := range selectors {
		element, err := waiter.FindElementRobust(selector)
		if err == nil {
			if err := element.Click(proto.InputMouseButtonLeft, 1); err == nil {
				log.Printf("成功点击登录确认按钮")
				return nil
			}
		}
	}
	
	return fmt.Errorf("无法找到登录确认按钮")
}