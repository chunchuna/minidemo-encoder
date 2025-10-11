@echo off
setlocal EnableDelayedExpansion

echo Current directory: "%CD%"
echo Searching for *.rec files in all subfolders...
echo.

set "count=0"

:: 递归搜索所有子目录的.rec文件
for /R %%d in (.) do (
    pushd "%%d"
    for %%f in (*.rec) do (
        if exist "%%f" (
            set /a count+=1
            set "random_name="
            for /L %%i in (1,1,8) do (
                set /a "digit=!random! %% 10"
                set "random_name=!random_name!!digit!"
            )
            
            echo [%%d] File: "%%f" will be renamed to "!random_name!player.rec"
            ren "%%f" "!random_name!player.rec"
            if !errorlevel! equ 0 (
                echo Success: Renamed "%%f" to "!random_name!player.rec"
            ) else (
                echo Failed to rename "%%f" - Error code: !errorlevel!
            )
        )
    )
    popd
)

echo.
if !count! equ 0 (
    echo No .rec files found in any directory.
    echo Please make sure:
    echo 1. Files have .rec extension (not .REC or similar)
    echo 2. You have the required permissions to access all folders
) else (
    echo Renamed !count! files successfully.
)

echo.
echo Renaming complete!
pause
