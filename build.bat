@echo off

@REM FOR /F "tokens=1" %%G IN ('git rev-list --max-parents=0 HEAD') DO SET result=%%G
@REM SET COMMIT=%result:~0,7%

@REM set NOW=%DATE% %TIME%

rmdir /s /q dist\release
mkdir dist\release
mkdir .tmp

mkdir .tmp\xgen\v0.9\dist
echo XGEN v0.9 > .tmp\xgen\v0.9\dist\index.html

rem Building XGEN v1.0
rem 在集成之前，需要修改xgen的环境变量BASE=__yao_admin_root，如果是前端单独测试，设置BASE=yao，或是清空BASE设置
rem
rem git clone git@github.com:wwsheng009/xgen.git ../xgen-v1.0

rem echo BASE=__yao_admin_root > ../xgen-v1.0/packages/xgen/.env
rem cd ../xgen-v1.0 && pnpm install --no-frozen-lockfile
rem set NODE_ENV=production
rem pnpm run build
rem echo BASE=yao > ../xgen-v1.0/packages/xgen/.env

rem Checkout init
rem del /s /q .tmp\yao-init
rem rmdir /s /q .tmp\yao-init
rem git clone https://github.com/wwsheng009/yao-init-0.10.3.git .tmp\yao-init
rem del /s /q .tmp\yao-init\.git
rem del /s /q .tmp\yao-init\.gitignore
rem del /s /q .tmp\yao-init\LICENSE
rem del /s /q .tmp\yao-init\README.md


rem Packing
mkdir .tmp\data\xgen
xcopy /e /y /q /i ui .tmp\data\ui
xcopy /e /y /q /i yao .tmp\data\yao
xcopy /e /y /q /i .tmp\xgen\v0.9\dist .tmp\data\xgen\v0.9
xcopy /e /y /q /i ..\xgen-v1.0\packages\xgen\dist .tmp\data\xgen\v1.0
xcopy /e /y /q /i ..\xgen-v1.0\packages\setup\build .tmp\data\xgen\setup
xcopy /e /y /q /i .tmp\yao-init .tmp\data\init
go-bindata -fs -pkg data -o data/bindata.go -prefix ".tmp/data/" .tmp/data/...
rmdir /s /q .tmp\data
rmdir /s /q .tmp\xgen

rem Replace PRVERSION
rem powershell -Command "(Get-Content share\const.go) | ForEach-Object { $_ -replace 'const PRVERSION = \"DEV\"', 'const PRVERSION = \"%COMMIT%-${NOW}-debug\"' } | Set-Content share\const.go"

rem Making artifacts
mkdir dist
rem del /q dist\release\yao-debug.exe
set CGO_ENABLED=1
set GOARCH=amd64
set GOOS=windows
go build -v -o dist\release\yao-debug.exe

del /Q %GOPATH%\bin\yao-dev.exe
move dist\release\yao-debug.exe %GOPATH%\bin\yao-dev.exe

rem Reset const

rem copy /Y share\const.goe share\const.go
rem del /Q share\const.goe