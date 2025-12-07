package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"
    "golang.org/x/sys/windows/registry"
    "path/filepath"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}


func (a *App) EnableAutoStart() error {
    exePath, err := os.Executable()
    if err != nil {
        return fmt.Errorf("failed to get executable path: %w", err)
    }

    fmt.Println(exePath)

    runKey, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
    if err != nil {
        return fmt.Errorf("failed to open registry key: %w", err)
    }
    defer runKey.Close()

    err = runKey.SetStringValue("YzuAutologin", exePath)
    if err != nil {
        return fmt.Errorf("failed to set registry value: %w", err)
    }

    return nil
}

func (a *App) DisableAutoStart() error {
    runKey, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
    if err != nil {
        return fmt.Errorf("failed to open registry key: %w", err)
    }
    defer runKey.Close()

    err = runKey.DeleteValue("YzuAutologin")
    if err != nil {
        return fmt.Errorf("failed to delete registry value: %w", err)
    }

    return nil
}


// SaveValue saves the given value to a JSON file
func (a *App) SaveValue(data map[string]string) error {
    file, err := os.Create("data.json")
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    if err := encoder.Encode(data); err != nil {
        return err
    }

    return nil
}

// ReadData reads the data from data.json file
func (a *App) ReadData() (map[string]string, error) {
    exePath, err := os.Executable()
    if (err != nil) {
        return nil, fmt.Errorf("failed to get executable path: %w", err)
    }
    exeDir := filepath.Dir(exePath)
    fmt.Println("Executable directory:", exeDir)
    filename := filepath.Join(exeDir, "data.json")
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    data := make(map[string]string)
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&data); err != nil {
        return nil, err
    }

    return data, nil
}

// DetectNetworkLoginPage 自动检测校园网登录页面
func (a *App) DetectNetworkLoginPage() (string, error) {
	detector, err := NewNetworkDetector(30 * time.Second)
	if err != nil {
		return "", fmt.Errorf("创建网络检测器失败: %w", err)
	}
	defer detector.Close()

	loginURL, err := detector.DetectLoginPage()
	if err != nil {
		return "", fmt.Errorf("检测登录页面失败: %w", err)
	}

	return loginURL, nil
}

// GetNetworkStatus 获取网络状态信息
func (a *App) GetNetworkStatus() (map[string]interface{}, error) {
	detector, err := NewNetworkDetector(30 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("创建网络检测器失败: %w", err)
	}
	defer detector.Close()

	status, err := detector.GetNetworkStatus()
	if err != nil {
		return nil, fmt.Errorf("获取网络状态失败: %w", err)
	}

	return status, nil
}

// AutoDetectAndSaveLoginURL 自动检测并保存登录URL
func (a *App) AutoDetectAndSaveLoginURL() (string, error) {
	loginURL, err := a.DetectNetworkLoginPage()
	if err != nil {
		return "", err
	}

	// 读取现有数据
	data, err := a.ReadData()
	if err != nil {
		// 如果文件不存在，创建新数据
		data = make(map[string]string)
	}

	// 更新登录URL
	data["webindex"] = loginURL

	// 保存数据
	err = a.SaveValue(data)
	if err != nil {
		return "", fmt.Errorf("保存登录URL失败: %w", err)
	}

	return loginURL, nil
}