@echo off

%1 %2
ver|find "5.">nul&&goto :Admin

setlocal enabledelayedexpansion
set Vbscript=Msgbox("卸载服务后，将无法使用工商银行薪资发放功能，是否确定卸载？",1,"系统管理员提醒")
for /f "Delims=" %%a in ('MsHta VBScript:Execute("CreateObject(""Scripting.Filesystemobject"").GetStandardStream(1).Write(%Vbscript:"=""%)"^)(Close^)') do Set "MsHtaReturnValue=%%a"
set ReturnValue1=确定
set ReturnValue2=取消或关闭窗口

if %MsHtaReturnValue% == 2  exit

echo 开始卸载服务.....
mshta vbscript:createobject("shell.application").shellexecute("%~s0","goto :Admin","","runas",1)(window.close)&goto :eof
:Admin
set path=%~dp0
set exe=%path%icbc_finance_service.exe
%exe% uninstall