@echo off

%1 %2
ver|find "5.">nul&&goto :Admin

setlocal enabledelayedexpansion
set Vbscript=Msgbox("ж�ط���󣬽��޷�ʹ�ù�������н�ʷ��Ź��ܣ��Ƿ�ȷ��ж�أ�",1,"ϵͳ����Ա����")
for /f "Delims=" %%a in ('MsHta VBScript:Execute("CreateObject(""Scripting.Filesystemobject"").GetStandardStream(1).Write(%Vbscript:"=""%)"^)(Close^)') do Set "MsHtaReturnValue=%%a"
set ReturnValue1=ȷ��
set ReturnValue2=ȡ����رմ���

if %MsHtaReturnValue% == 2  exit

echo ��ʼж�ط���.....
mshta vbscript:createobject("shell.application").shellexecute("%~s0","goto :Admin","","runas",1)(window.close)&goto :eof
:Admin
set path=%~dp0
set exe=%path%icbc_finance_service.exe
%exe% uninstall