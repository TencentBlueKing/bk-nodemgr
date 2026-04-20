@echo off
setlocal EnableDelayedExpansion
set prog_name=%1
set cu_date=%date:~0,4%-%date:~5,2%-%date:~8,2%
set cu_time=%time:~0,8%
set script_dir=%~dp0
cd /d %script_dir%

rem 判断进程是否已经存在（精确匹配路径）
set "expected_path=%script_dir%!prog_name!.exe"
powershell -command "$expectedPath = '%script_dir%!prog_name!.exe' -replace '/', '\'; $process = Get-Process -Name '!prog_name!' -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $expectedPath }; if ($process) { exit 0 } else { exit 1 }" >nul 2>&1
if %errorlevel% equ 0 (
    echo [%cu_date% %cu_time%] !prog_name! already exist in current directory
    goto EOF
)

rem 启动进程
start /b ./"!prog_name!.exe" -c ../etc/!prog_name!.conf >nul
ping -n 2 127.0.0.1 >nul 2>&1
powershell -command "$expectedPath = '%script_dir%!prog_name!.exe' -replace '/', '\'; $process = Get-Process -Name '!prog_name!' -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $expectedPath }; if ($process) { exit 0 } else { exit 1 }" >nul 2>&1
if %errorlevel% neq 0 (
    echo [%cu_date% %cu_time%] start !prog_name! fail
    goto EOF
)

echo [%cu_date% %cu_time%] start !prog_name! done

:EOF
