@echo off
for /r %%i in (go.sum) do (
    echo. > "%%i"
    echo Cleared "%%i"
)
echo All go.sum files have been cleared.
pause
