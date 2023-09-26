@echo off
setlocal enabledelayedexpansion

echo Checking Go version...
go version

for /r %%i in (go.mod) do (
    echo Processing directory: %%~dpi
    cd %%~dpi
    go mod tidy
)

endlocal
