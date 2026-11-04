@echo off
cd /d "%~dp0"
echo 正在构建，首次构建需要下载依赖，可能耗时几分钟...
echo 注意：这里不使用 -clean，避免误删已生成的 account.json。
wails build -s
if errorlevel 1 (
  echo.
  echo 构建失败。请先执行 wails doctor 检查环境。
  pause
  exit /b 1
)
echo.
echo 构建完成：build\bin\TextCrypto.exe
pause
