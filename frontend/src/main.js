import './style.css';

import {SaveValue, ReadData, Loginyzu, EnableAutoStart, DisableAutoStart, TestConnection, DetectNetworkLoginPage, AutoDetectAndSaveLoginURL, GetNetworkStatus} from '../wailsjs/go/main/App';
import { Quit } from '../wailsjs/runtime/runtime';
import 'sober';

let webindex = document.getElementById("webindex");
let countindex = document.getElementById("countindex");
let passwordindex = document.getElementById("passwordindex");
let operatorindex = document.getElementById("operatorindex");
let autostartindex = document.getElementById("autostartindex");
let operatorindex_items = document.querySelectorAll('s-segmented-button-item');
let testconnectindex = document.getElementById("testconnectindex");
let detectLoginPageBtn = document.getElementById("detectLoginPage");

// 设置一个定时器变量
let typingTimer;
let doneTypingInterval = 500; // 时间间隔（毫秒）

// 用户停止输入后的处理函数
function doneTyping() {

    try {
        SaveValue({
            [webindex.id]: webindex.value,
            [countindex.id]: countindex.value,
            [passwordindex.id]: passwordindex.value,
            [operatorindex.id]: operatorindex.value,
            [autostartindex.id]: autostartindex.checked.toString()
        })
    } catch (err) {
        console.error(err); 
    }
}

// 在用户输入时清除定时器
webindex.addEventListener('input', () => {
    clearTimeout(typingTimer);
    typingTimer = setTimeout(doneTyping, doneTypingInterval);
});

countindex.addEventListener('input', () => {
    clearTimeout(typingTimer);
    typingTimer = setTimeout(doneTyping, doneTypingInterval);
});

passwordindex.addEventListener('input', () => {
    clearTimeout(typingTimer);
    typingTimer = setTimeout(doneTyping, doneTypingInterval);
});

operatorindex.addEventListener('click', () => {
    clearTimeout(typingTimer);
    typingTimer = setTimeout(doneTyping, doneTypingInterval);
});

autostartindex.addEventListener('click', () => {
    clearTimeout(typingTimer);
    typingTimer = setTimeout(doneTyping, doneTypingInterval);
    try {
        if (autostartindex.checked) {
            EnableAutoStart();
        } else {
            DisableAutoStart();
        }
    } catch (err) {
        console.error('Failed to enable auto start:', err);
    }
}); 

testconnectindex.addEventListener('click', async () => {
    try {
        const result = await TestConnection();
        showSnackbar("连接测试完成");
        console.log("连接测试结果:");
        console.log(result);
        
        // 可选：将结果显示在页面上
        showDetailedResult(result);
    } catch (err) {
        console.error(err);
        showSnackbar("连接测试失败: " + err.toString());
    }
});

// 添加实际登录按钮功能
testconnectindex.addEventListener('contextmenu', async (e) => {
    e.preventDefault(); // 阻止右键菜单
    try {
        await Loginyzu();
        showSnackbar("正在登录...");
    } catch (err) {
        console.error(err);
        showSnackbar("登录失败: " + err.toString());
    }
});

// 添加工具提示
detectLoginPageBtn.title = "自动检测校园网登录页面，无需手动输入URL";
testconnectindex.title = "左键：测试网络连接 | 右键：执行自动登录";

// 自动检测登录页面功能
detectLoginPageBtn.addEventListener('click', async () => {
    try {
        showSnackbar("正在检测校园网登录页面，请稍候...");
        
        // 显示加载状态
        detectLoginPageBtn.disabled = true;
        detectLoginPageBtn.textContent = "检测中...";
        
        const loginURL = await AutoDetectAndSaveLoginURL();
        
        // 更新输入框
        webindex.value = loginURL;
        
        // 触发保存
        doneTyping();
        
        showSnackbar(`成功检测到登录页面: ${loginURL}`);
        
        // 显示网络状态信息
        const status = await GetNetworkStatus();
        showNetworkStatus(status);
        
    } catch (err) {
        console.error(err);
        showSnackbar("检测失败: " + err.toString());
    } finally {
        // 恢复按钮状态
        detectLoginPageBtn.disabled = false;
        detectLoginPageBtn.textContent = "自动检测";
    }
});

// 显示网络状态信息
function showNetworkStatus(status) {
    const statusDiv = document.createElement('div');
    statusDiv.style.cssText = `
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        background: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        z-index: 1000;
        max-width: 500px;
        max-height: 80vh;
        overflow-y: auto;
    `;
    
    let statusHTML = `<h3>网络状态信息</h3>`;
    
    if (status.connected) {
        statusHTML += `<p style="color: green;">✅ ${status.connectivity_result}</p>`;
    } else {
        statusHTML += `<p style="color: orange;">⚠️ ${status.connectivity_result}</p>`;
        
        if (status.needs_authentication && status.login_url) {
            statusHTML += `<p style="color: blue;">🔗 检测到登录页面: ${status.login_url}</p>`;
        } else if (status.detection_error) {
            statusHTML += `<p style="color: red;">❌ 检测错误: ${status.detection_error}</p>`;
        }
    }
    
    // 显示原始状态数据（调试用）
    statusHTML += `<hr><details><summary>详细数据</summary><pre style="font-size: 12px; overflow: auto;">${JSON.stringify(status, null, 2)}</pre></details>`;
    
    statusHTML += `<button id="closeStatus" style="
        margin-top: 10px;
        padding: 8px 16px;
        background: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
    ">关闭</button>`;
    
    statusDiv.innerHTML = statusHTML;
    document.body.appendChild(statusDiv);
    
    // 添加关闭按钮事件
    document.getElementById('closeStatus').addEventListener('click', () => {
        document.body.removeChild(statusDiv);
    });
    
    // 点击外部关闭
    statusDiv.addEventListener('click', (e) => {
        if (e.target === statusDiv) {
            document.body.removeChild(statusDiv);
        }
    });
    
    // 10秒后自动关闭
    setTimeout(() => {
        if (document.body.contains(statusDiv)) {
            document.body.removeChild(statusDiv);
        }
    }, 10000);
}

// 显示详细测试结果的函数
function showDetailedResult(result) {
    // 创建一个模态对话框或通知区域显示结果
    const resultDiv = document.createElement('div');
    resultDiv.style.cssText = `
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        background: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        z-index: 1000;
        max-width: 400px;
        white-space: pre-line;
    `;
    
    resultDiv.innerHTML = `
        <h3>连接测试结果</h3>
        <p>${result}</p>
        <button id="closeResult" style="
            margin-top: 10px;
            padding: 8px 16px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        ">关闭</button>
    `;
    
    document.body.appendChild(resultDiv);
    
    // 添加关闭按钮事件
    document.getElementById('closeResult').addEventListener('click', () => {
        document.body.removeChild(resultDiv);
    });
    
    // 5秒后自动关闭
    setTimeout(() => {
        if (document.body.contains(resultDiv)) {
            document.body.removeChild(resultDiv);
        }
    }, 5000);
}

ReadData()
    .then((data) => {
        webindex.value = data.webindex;
        countindex.value = data.countindex;
        passwordindex.value = data.passwordindex;
        operatorindex.value = data.operatorindex;
        autostartindex.checked = data.autostartindex == "true";
        
        if (data.autostartindex == "true") {
            (async () => {
                try {
                    await Loginyzu();
                    showSnackbar("测试连接已触发");
                } catch (err) {
                    console.error(err);
                    showSnackbar(err.toString());
                }
            })();
        }

        setTimeout(() => {
            operatorindex_items.forEach((item) => {
                forceRedraw(item);
            });
        }, 1000);
        
    })
    .catch((err) => {
        console.error("Error reading data:", err);
    }
);

let exitelement = document.getElementById("exit");

exitelement.addEventListener('click', () => {
    try {
        Quit();
    } catch (err) {
        console.error(err); 
    }
});

// 强制重绘 s-segmented-button 元素
function forceRedraw(element) {
    element.style.display = 'none';
    element.offsetHeight; // 触发重绘
    element.style.display = '';
};

// 显示 Snackbar 消息通知
function showSnackbar(message) {
    const snackbar = document.createElement('s-snackbar');
    const htmlContent = `
        <s-button slot="trigger" class="s-button--text" style="background-color: transparent"></s-button>
        ${message}
    `;
    snackbar.innerHTML = htmlContent; 
    
    const sPage = document.querySelector('s-page');
    sPage.appendChild(snackbar);

    snackbar.querySelector('s-button').click();
    setTimeout(() => {
        snackbar.remove();
    }, 5000); 
}
