package main

import (
	"fmt"
	"log"
	"strings"
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
	
	// 减少额外等待时间，使用更智能的检测
	time.Sleep(500 * time.Millisecond)
	
	// 快速检查页面是否包含关键元素
	checkSelectors := []string{
		"input", "form", "button",
		"input[type='text']", "input[type='password']",
		"input[name*='user']", "input[name*='pass']",
	}
	
	foundElements := false
	for _, selector := range checkSelectors {
		elements, err := sw.page.Elements(selector)
		if err == nil && len(elements) > 0 {
			if !foundElements {
				log.Printf("页面加载完成，找到元素: %s (数量: %d)", selector, len(elements))
				foundElements = true
			}
			// 找到关键元素就立即返回，不检查所有选择器
			if strings.Contains(selector, "input") {
				return nil
			}
		}
	}
	
	if !foundElements {
		log.Printf("页面已加载，但未找到关键表单元素")
	}
	return nil
}

// FindElementRobust 健壮的元素查找
func (sw *SmartWaiter) FindElementRobust(selector ElementSelector) (*rod.Element, error) {
	// 增加重试机制，但减少重试间隔
	maxRetries := 2  // 减少重试次数
	var lastErr error
	
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("元素查找重试 %d/%d", retry, maxRetries)
			time.Sleep(200 * time.Millisecond)  // 减少重试间隔
		}
		
		// 尝试主要选择器
		if selector.Primary != "" {
			element, err := sw.page.Element(selector.Primary)
			if err == nil {
				log.Printf("通过主要选择器找到元素: %s", selector.Primary)
				return element, nil
			}
			lastErr = err
		}
		
		// 尝试备选选择器
		for _, alt := range selector.Alternatives {
			element, err := sw.page.Element(alt)
			if err == nil {
				log.Printf("通过备选选择器找到元素: %s", alt)
				return element, nil
			}
		}
		
		// 尝试通过属性查找
		for _, attr := range selector.Attributes {
			elements, err := sw.page.Elements(fmt.Sprintf("[%s]", attr))
			if err == nil && len(elements) > 0 {
				log.Printf("通过属性找到元素: [%s] (数量: %d)", attr, len(elements))
				// 返回第一个可见的元素
				for _, element := range elements {
					if visible, _ := element.Visible(); visible {
						return element, nil
					}
				}
				return elements[0], nil
			}
		}
		
		// 尝试通过文本内容查找
		for _, text := range selector.TextContains {
			element, err := sw.page.Element(fmt.Sprintf("//*[contains(text(), '%s')]", text))
			if err == nil {
				log.Printf("通过文本内容找到元素: 包含 '%s'", text)
				return element, nil
			}
		}
		
		// 尝试模糊匹配
		if retry == maxRetries-1 {
			// 最后一次尝试：查找所有输入框
			elements, err := sw.page.Elements("input")
			if err == nil && len(elements) > 0 {
				log.Printf("找到 %d 个输入框，尝试匹配", len(elements))
				for _, element := range elements {
					// 检查元素属性是否匹配
					for _, attr := range selector.Attributes {
						attrValue, _ := element.Attribute(attr)
						if attrValue != nil {
							for _, text := range selector.TextContains {
								if strings.Contains(strings.ToLower(*attrValue), strings.ToLower(text)) {
									log.Printf("通过模糊匹配找到元素: %s='%s'", attr, *attrValue)
									return element, nil
								}
							}
						}
					}
				}
			}
		}
	}
	
	return nil, fmt.Errorf("element not found with any selector: %w", lastErr)
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
			MaxRetries:  2,  // 减少重试次数
			Timeout:     8 * time.Second,  // 减少超时时间
		},
		{
			Name:        "输入用户名",
			Description: "在用户名输入框中输入账号",
			Execute:     inputUsername,
			MaxRetries:  2,  // 减少重试次数
			Timeout:     3 * time.Second,  // 减少超时时间
		},
		{
			Name:        "输入密码",
			Description: "在密码输入框中输入密码",
			Execute:     inputPassword,
			MaxRetries:  2,  // 减少重试次数
			Timeout:     3 * time.Second,  // 减少超时时间
		},
		{
			Name:        "提交表单",
			Description: "提交登录表单",
			Execute:     submitForm,
			MaxRetries:  2,
			Timeout:     3 * time.Second,  // 减少超时时间
		},
		{
			Name:        "选择运营商",
			Description: "选择网络运营商",
			Execute:     selectOperator,
			MaxRetries:  2,  // 减少重试次数
			Timeout:     3 * time.Second,  // 减少超时时间
		},
		{
			Name:        "确认登录",
			Description: "点击最终登录按钮",
			Execute:     confirmLogin,
			MaxRetries:  2,  // 减少重试次数
			Timeout:     3 * time.Second,  // 减少超时时间
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
	
	// 减少额外等待时间
	time.Sleep(300 * time.Millisecond)
	
	// 定义多种可能的选择器
	selectors := []ElementSelector{
		{
			Primary:      "input[name='username']",
			Alternatives: []string{
				"input[name='username_tip']", 
				"input[type='text']", 
				"input[id*='username']", 
				"input[class*='username']",
				"input[placeholder*='用户']",
				"input[placeholder*='账号']",
				"input[placeholder*='学号']",
				"input[placeholder*='Username']",
				"input[placeholder*='Account']",
				"#username",
				".username",
				"input:text",
			},
			Attributes:   []string{"placeholder", "name", "id", "class", "type"},
			TextContains: []string{"用户名", "账号", "学号", "username", "account", "user"},
		},
	}
	
	// 先尝试查找所有文本输入框
	elements, err := page.Elements("input[type='text']")
	if err == nil && len(elements) > 0 {
		log.Printf("找到 %d 个文本输入框，尝试第一个", len(elements))
		element := elements[0]
		
		// 检查元素是否可见和可点击
		if err := waitForElementReady(element); err != nil {
			log.Printf("第一个文本输入框不可用: %v", err)
		} else {
			return fillInputElement(element, config.Countindex, "用户名")
		}
	}
	
	for _, selector := range selectors {
		element, err := waiter.FindElementRobust(selector)
		if err == nil {
			return fillInputElement(element, config.Countindex, "用户名")
		}
	}
	
	// 最后尝试：查找所有输入框，排除密码框
	allInputs, err := page.Elements("input")
	if err == nil {
		for _, input := range allInputs {
			inputType, _ := input.Attribute("type")
			if inputType != nil && *inputType == "password" {
				continue // 跳过密码框
			}
			
			if err := fillInputElement(input, config.Countindex, "用户名"); err == nil {
				return nil
			}
		}
	}
	
	return fmt.Errorf("找不到用户名输入框")
}

// waitForElementReady 等待元素准备好
func waitForElementReady(element *rod.Element) error {
	// 检查元素是否可见
	visible, err := element.Visible()
	if err != nil || !visible {
		return fmt.Errorf("元素不可见")
	}
	
	// 检查元素是否启用
	disabled, err := element.Attribute("disabled")
	if err == nil && disabled != nil && *disabled == "true" {
		return fmt.Errorf("元素被禁用")
	}
	
	// 减少等待时间
	time.Sleep(100 * time.Millisecond)
	return nil
}

// fillInputElement 填充输入框的通用函数
func fillInputElement(element *rod.Element, value string, fieldName string) error {
	// 等待元素准备好
	if err := waitForElementReady(element); err != nil {
		return fmt.Errorf("%s输入框未准备好: %w", fieldName, err)
	}
	
	// 点击输入框
	if err := element.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("点击%s输入框失败: %w", fieldName, err)
	}
	
	// 全选并清空
	if err := element.SelectAllText(); err != nil {
		log.Printf("无法选择%s文本，尝试直接输入: %v", fieldName, err)
		// 如果选择失败，尝试直接输入
	}
	
	if err := element.Input(""); err != nil {
		return fmt.Errorf("清空%s输入框失败: %w", fieldName, err)
	}
	
	// 输入值
	if err := element.Input(value); err != nil {
		return fmt.Errorf("输入%s失败: %w", fieldName, err)
	}
	
	log.Printf("成功输入%s", fieldName)
	return nil
}

// inputPassword 输入密码
func inputPassword(page *rod.Page, config *Config) error {
	waiter := NewSmartWaiter(page)
	
	// 减少额外等待时间
	time.Sleep(200 * time.Millisecond)
	
	// 定义多种可能的选择器
	selectors := []ElementSelector{
		{
			Primary:      "input[type='password']",
			Alternatives: []string{
				"input[name='password']", 
				"input[name='pwd_tip']", 
				"input[id*='password']", 
				"input[class*='password']",
				"input[placeholder*='密码']",
				"input[placeholder*='Password']",
				"input[placeholder*='pass']",
				"#password",
				".password",
			},
			Attributes:   []string{"type", "name", "id", "class", "placeholder"},
			TextContains: []string{"密码", "password", "pass"},
		},
	}
	
	// 先尝试查找所有密码输入框
	elements, err := page.Elements("input[type='password']")
	if err == nil && len(elements) > 0 {
		log.Printf("找到 %d 个密码输入框，尝试第一个", len(elements))
		element := elements[0]
		return fillInputElement(element, config.Passwordindex, "密码")
	}
	
	for _, selector := range selectors {
		element, err := waiter.FindElementRobust(selector)
		if err == nil {
			return fillInputElement(element, config.Passwordindex, "密码")
		}
	}
	
	// 最后尝试：查找所有输入框，找到密码类型的
	allInputs, err := page.Elements("input")
	if err == nil {
		for _, input := range allInputs {
			inputType, _ := input.Attribute("type")
			if inputType != nil && *inputType == "password" {
				if err := fillInputElement(input, config.Passwordindex, "密码"); err == nil {
					return nil
				}
			}
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
	
	// 减少等待时间
	time.Sleep(1 * time.Second)
	
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
	
	// 减少等待下拉选项加载时间
	time.Sleep(500 * time.Millisecond)
	
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
	
	// 减少等待时间
	time.Sleep(500 * time.Millisecond)
	
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