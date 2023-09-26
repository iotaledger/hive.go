@echo off
setlocal enabledelayedexpansion

REM Set paths and tools
set ROOT_DIR=C:\Users\the\Documents\node\zipp.foundation-master
set GOPATH_BIN=%GOPATH%\bin
set PROTOC_GEN_GO=%GOPATH_BIN%\protoc-gen-go.exe

REM Install protoc-gen-go if it doesn't exist
if not exist "%PROTOC_GEN_GO%" (
    echo Installing protoc-gen-go...
    go get -u google.golang.org/protobuf/cmd/protoc-gen-go
)

REM Compile all proto files to pb.go
for /r %%f in (*.proto) do (
    echo Current Directory: %%~dpf
    set RELATIVE_PATH=%%f
    set RELATIVE_PATH=!RELATIVE_PATH:%ROOT_DIR%\=!
    REM Convert \ to /
    set RELATIVE_PATH=!RELATIVE_PATH:\=/!
    echo Compiling !RELATIVE_PATH!...
    cd /d %%~dpf
    protoc --proto_path=%ROOT_DIR% --plugin=protoc-gen-go=%PROTOC_GEN_GO% --go_out=paths=source_relative:. !RELATIVE_PATH!
    cd /d %ROOT_DIR%
)

echo Done!
pause
