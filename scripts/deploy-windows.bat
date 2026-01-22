@echo off
REM AgentFlow Windows 快速部署脚本
REM 用于快速部署 Claude 开发环境

setlocal EnableDelayedExpansion

REM 颜色代码（Windows 10+ 支持）
set "INFO=[94m"
set "SUCCESS=[92m"
set "WARNING=[93m"
set "ERROR=[91m"
set "NC=[0m"

REM 打印带颜色的消息
goto :main

:print_header
echo.
echo %CYAN%━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%NC%
echo %CYAN%%1%NC%
echo %CYAN%━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%NC%
echo.
goto :eof

:print_info
echo %INFO%ℹ️  %~1%NC%
goto :eof

:print_success
echo %SUCCESS%✅ %~1%NC%
goto :eof

:print_warning
echo %WARNING%⚠️  %~1%NC%
goto :eof

:print_error
echo %ERROR%❌ %~1%NC%
goto :eof

REM 检查 Windows 环境
:check_windows_environment
call :print_header "检查 Windows 环境"

for /f "tokens=4-5 delims=. " %%i in ('ver') do set WINDOWS_VERSION=%%i
echo %SUCCESS%Windows 版本: %WINDOWS_VERSION%

REM 检查架构
if defined PROCESSOR_ARCHITECTURE (
    echo %INFO%系统架构: %PROCESSOR_ARCHITECTURE%
)

REM 检查 Chocolatey
where choco >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo %SUCCESS%Chocolatey 已安装
) else (
    echo %WARNING%Chocolatey 未安装
    echo %INFO%推荐安装 Chocolatey 以简化依赖管理
    echo %INFO%安装命令:
    echo     powershell -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol::Tls12, Tls13, Tls14, Tls15, Tls16; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
    echo.
    pause
)

goto :eof

REM 安装 Claude CLI
:install_claude_cli
call :print_header "安装 Claude CLI"

where claudecli >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    for /f "tokens=*" %%i in ('claudecli --version') do set CLAUDE_VERSION=%%i
    echo %SUCCESS%Claude CLI 已安装: !CLAUDE_VERSION!
    echo %INFO%Claude CLI 路径:
    where claudecli
) else (
    echo %INFO%Claude CLI 未安装，开始安装...

    REM 检查 npm
    where npm >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo %INFO%使用 npm 安装 Claude CLI...
        call npm install -g @anthropic-ai/claude-cli
        if %ERRORLEVEL% EQU 0 (
            echo %SUCCESS%Claude CLI 安装完成
        ) else (
            echo %ERROR%npm 安装失败
            echo %INFO%请手动安装 Claude CLI:
            echo     方法 1: npm install -g @anthropic-ai/claude-cli
            echo     方法 2: 下载二进制: https://github.com/anthropics/claude-cli/releases
            pause
            exit /b 1
        )
    ) else (
        echo %ERROR%npm 未安装
        echo %INFO%请先安装 Node.js 和 npm
        echo %INFO%推荐使用 Chocolatey 安装: choco install nodejs
        pause
        exit /b 1
    )
)

goto :eof

REM 安装 Python 依赖
:install_python_dependencies
call :print_header "安装 Python 依赖"

where python >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    for /f "tokens=*" %%i in ('python --version') do set PYTHON_VERSION=%%i
    echo %SUCCESS%Python 已安装: %PYTHON_VERSION%

    REM 安装 Python 依赖
    if exist "python\requirements.txt" (
        echo %INFO%安装 Python 依赖...
        cd python
        pip install -r requirements.txt
        cd ..
        echo %SUCCESS%Python 依赖安装完成
    )
) else (
    echo %WARNING%Python 未安装
    echo %INFO%请安装 Python 3.8+
    echo     推荐方式: 访问 https://www.python.org/downloads/
    echo     或使用: winget install Python.Python.3.11
    pause
)

goto :eof

REM 安装 Go 依赖
:install_go_dependencies
call :print_header "安装 Go 依赖"

where go >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
    echo %SUCCESS%Go 已安装: %GO_VERSION%

    REM 设置 Go 代理（中国）
    set GOPROXY=https://goproxy.cn,direct
    echo %INFO%Go 代理已设置: %GOPROXY%

    REM 安装依赖
    if exist "golang" (
        echo %INFO%安装 Go 模块依赖...
        cd golang
        go mod tidy
        cd ..
        echo %SUCCESS%Go 依赖安装完成
    )
) else (
    echo %WARNING%Go 未安装
    echo %INFO%请安装 Go 1.21+
    echo     推荐方式1: 下载官方包: https://go.dev/dl/
    echo     推荐方式2: 使用 winget: winget install GoLang.Go
    pause
)

goto :eof

REM 编译 Go 二进制文件
:build_go_binaries
call :print_header "编译 Go 二进制文件"

if not exist "golang" (
    call :print_error "golang 目录不存在"
    exit /b 1
)

cd golang

REM 创建输出目录
if not exist "bin" mkdir bin

REM 编译 master
echo %INFO%编译 master...
go build -o bin/master cmd/master/main.go
if %ERRORLEVEL% NEQ 0 (
    call :print_error "编译 master 失败"
    cd ..
    exit /b 1
)

REM 编译 worker
echo %INFO%编译 worker...
go build -o bin/worker cmd/worker/main.go
if %ERRORLEVEL% NEQ 0 (
    call :print_error "编译 worker 失败"
    cd ..
    exit /b 1
)

REM 编译 oneshot
echo %INFO%编译 oneshot...
go build -o bin/oneshot cmd/oneshot/main.go
if %ERRORLEVEL% NEQ 0 (
    call :print_error "编译 oneshot 失败"
    cd ..
    exit /b 1
)

cd ..

call :print_success "编译完成"
echo %INFO%二进制文件位置:
echo     - golang\bin\master.exe
echo     - golang\bin\worker.exe
echo     - golang\bin\oneshot.exe

goto :eof

REM 配置文件边界
:setup_file_boundaries
call :print_header "配置文件边界"

set "BOUNDARIES_FILE=.agentflow\boundaries.json"

if exist "%BOUNDARIES_FILE%" (
    call :print_success "文件边界配置已存在: %BOUNDARIES_FILE%"
) else (
    echo %INFO%创建文件边界配置...
    mkdir .agentflow 2>nul

    if exist ".agentflow\boundaries.example.json" (
        copy ".agentflow\boundaries.example.json" "%BOUNDARIES_FILE%" >nul
    ) else (
        echo %INFO%创建默认配置...
        (
            echo {
            echo   "agent-frontend": [
            echo     {
            echo       "file_pattern": "src/frontend/**/*",
            echo       "access_type": "exclusive",
            echo       "description": "Frontend agent manages frontend UI code"
            echo     },
            echo     {
            echo       "file_pattern": "src/api/**/*",
            echo       "access_type": "readonly",
            echo       "description": "Frontend agent can read API definitions"
            echo     }
            echo   ],
            echo   "agent-backend": [
            echo     {
            echo       "file_pattern": "src/backend/**/*",
            echo       "access_type": "exclusive",
            echo       "description": "Backend agent manages backend code"
            echo     },
            echo     {
            echo       "file_pattern": "src/api/**/*",
            echo       "access_type": "shared",
            echo       "description": "Backend agent shares API files"
            echo     }
            echo   ],
            echo   "agent-database": [
            echo     {
            echo       "file_pattern": "src/database/**/*",
            echo       "access_type": "exclusive",
            echo       "description": "Database agent manages database layer"
            echo     }
            echo   ]
            echo }
        ) > "%BOUNDARIES_FILE%"
    )

    call :print_success "文件边界配置已创建: %BOUNDARIES_FILE%"
)

goto :eof

REM 启动 Master 服务
:start_master
call :print_header "启动 Master 服务"

REM 检查是否已运行
tasklist /FI "IMAGENAME eq golang\bin\master.exe" 2>nul | findstr "." >nul
if %ERRORLEVEL% EQU 0 (
    call :print_warning "Master 已在运行"
    goto :eof
)

if exist "golang\bin\master.exe" (
    start "" /b golang\bin\master.exe -config golang\config.example.yaml
    call :print_success "Master 已启动（新窗口）"
    echo %INFO%Master 地址: http://localhost:8848
) else (
    call :print_warning "未找到 master.exe，尝试 Python 版本..."
    if exist "python\agentflow\__init__.py" (
        start python\agentflow-cli master --port 8848
        call :print_success "Master 已启动（Python 版本）"
    ) else (
        call :print_error "未找到可执行的 master"
    )
)

goto :eof

REM 启动 Worker
:start_worker
call :print_header "启动 Worker"

set "WORKER_GROUP=%~1"
if "%WORKER_GROUP%"=="" set "WORKER_GROUP=default"

if exist "golang\bin\worker.exe" (
    start "" /b golang\bin\worker.exe -config golang\config.example.yaml
    call :print_success "Worker 已启动（组: %WORKER_GROUP%）"
) else (
    call :print_warning "未找到 worker.exe，尝试 Python 版本..."
    start python\agentflow-cli worker --group %WORKER_GROUP% --auto
    call :print_success "Worker 已启动（Python 版本）"
)

goto :eof

REM 停止服务
:stop_services
call :print_header "停止服务"

REM 停止 Master
tasklist /FI "IMAGENAME eq golang\bin\master.exe" 2>nul | findstr "." >nul
if %ERRORLEVEL% EQU 0 (
    echo %INFO%停止 Master...
    taskkill /F /IM golang\bin\master.exe >nul 2>&1
    call :print_success "Master 已停止"
)

REM 停止所有 Workers
tasklist /FI "IMAGENAME eq golang\bin\worker.exe" 2>nul | findstr "." >nul
if %ERRORLEVEL%EQU 0 (
    echo %INFO%停止 Workers...
    taskkill /F /IM golang\bin\worker.exe >nul 2>&1
    call :print_success "Workers 已停止"
)

REM 停止 Python 版本
taskkill /F /IM python.exe /FI "WINDOWTITLE eq agentflow*" >nul 2>&1

goto :eof

REM 状态检查
:show_status
call :print_header "系统状态"

REM 检查 Master
tasklist /FI "IMAGENAME eq golang\bin\master.exe" 2>nul | findstr "." >nul
if %ERRORLEVEL% EQU 0 (
    call :print_success "Master 运行中"
) else (
    call :print_warning "Master 未运行"
)

REM 检查 Workers
tasklist /FI "IMAGENAME eq golang\bin\worker.exe" 2>nul | findstr "." >nul
if %ERRORLEVEL EQU 0 (
    for /f %%c in ('tasklist /FI "IMAGENAME eq golang\bin\worker.exe" 2^>nul^| find "." /c') do (
    set /a COUNT+=1
)
    call :print_success "Workers 运行中: !COUNT! 个
) else (
    call :print_warning "没有运行的 Workers"
)

REM 检查二进制文件
if exist "golang\bin\master.exe" (
    call :print_success "Go master: golang\bin\master.exe"
)
if exist "golang\bin\worker.exe" (
    call :print_success "Go worker: golang\bin\worker.exe"
)
if exist "golang\bin\oneshot.exe" (
    call :print_success "Go oneshot: golang\bin\oneshot.exe"
)

REM 检查配置文件
if exist ".agentflow\boundaries.json" (
    call :print_success "文件边界配置: .agentflow\boundaries.json"
)

goto :eof

REM 显示使用指南
:show_usage
echo.
echo %SUCCESS%AgentFlow Windows 快速部署完成！%NC%
echo.
echo 📝 常用命令:
echo.
echo   %YELLOW%# 查看系统状态%NC%
echo   scripts\deploy-windows.bat status
echo.
echo   %YELLOW%# 创建任务%NC%
   scripts\quick-task.bat "测试任务" "shell:echo Hello World"
echo.
echo   %YELLOW%# 停止服务%NC%
echo   scripts\deploy-windows.bat stop
echo.
echo 🔗 服务地址:
echo   - Master API: http://localhost:8848
echo   - API 文档: http://localhost:8848/docs
echo.
echo 📚 文档:
echo   - 完整指南: docs\git-integration-guide.md
echo   - README.md: README.md
echo   - Skill 手册: skills\agentflow.md
echo.
echo 🎯 下一步:
echo    1. 访问 Master API: http://localhost:8848
echo    2. 创建第一个任务测试系统
echo    3. 配置文件边界: .agentflow\boundaries.json
echo    4. 在新窗口启动多个 Workers 并行处理
echo.

goto :eof

REM 检查环境
:check
call :check_windows_environment
echo.
goto :eof

REM 安装所有依赖
:install
call :check_windows_environment
call :install_claude_cli
call :install_python_dependencies
call :install_go_dependencies
echo.
call :print_success "所有依赖已安装！"
goto :eof

REM 构建二进制
:build
call :build_go_binaries
call :setup_file_boundaries
echo.
call :print_success "构建完成！"
goto :eof

REM 完整部署
:deploy
call :print_header "AgentFlow Windows 快速部署"

call :check
call :install
call :build
call :start_master

REM 等待用户按键
echo.
pause
call :show_usage
goto :eof

REM 主函数
:main
if "%~1"=="" goto :deploy

if "%~1"=="check" goto :check
if "%~1"=="install" goto :install
if "%~1"=="build" goto :build
if "%~1"=="start" goto :start_master
if "%~1"=="worker" goto :start_worker
if "%~1"=="stop" goto :stop_services
if "%~1"=="status" goto :show_status

echo AgentFlow Windows 快速部署脚本
echo.
echo 用法: %~nx0 ^<命令^>
echo.
echo 命令:
echo   check      - 检查环境
echo   install    - 安装所有依赖
echo   build      - 编译二进制文件
echo   deploy     - 完整部署（默认）
echo   start      - 启动 Master
echo   worker [组] - 启动 Worker（可选组名）
echo   stop       - 停止所有服务
echo   status     - 查看系统状态
echo.
echo 示例:
echo   %~nx0 deploy          # 完整部署
echo   %~nx0 status          # 查看状态
echo   %~nx0 worker backend   # 启动 backend 组 Worker
echo.
goto :eof
