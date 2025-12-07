import './style.css';

import {SaveValue, ReadData, Loginyzu, EnableAutoStart, DisableAutoStart, TestConnection} from '../wailsjs/go/main/App';
import { Quit } from '../wailsjs/runtime/runtime';
import 'sober';

let webindex = document.getElementById("webindex");
let countindex = document.getElementById("countindex");
let passwordindex = document.getElementById("passwordindex");
let operatorindex = document.getElementById("operatorindex");
let autostartindex = document.getElementById("autostartindex");
let operatorindex_items = document.querySelectorAll('s-segmented-button-item');
let testconnectindex = document.getElementById("testconnectindex");

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
        showSnackbar("连接测试完成，请查看控制台");
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
        showSnackbar("正在执行自动登录...");
    } catch (err) {
        console.error(err);
        showSnackbar("登录失败: " + err.toString());
    }
});

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
