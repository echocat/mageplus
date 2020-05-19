@ECHO OFF
SETLOCAL
REM ##############################################################################
REM ##                                                                          ##
REM ##  mageplus bootstrap wrapper for Windows systems                          ##
REM ##                                                                          ##
REM ##############################################################################
REM ##  DO NOT EDIT!!                                                           ##
REM ##############################################################################

SET dirName=%~dp0
SET os=Windows
SET arch=32bit
SET ext=.exe
SET downloadExt=.zip
IF "%PROCESSOR_ARCHITECTURE%" == "AMD64" (
    SET arch=64bit
)

IF NOT EXIST "%dirName%\mageplusw" (
    CALL :fatal This mageplus wrapper was not initiated correctly. Try download mageplus binary and run: mageplus -wrapper
    EXIT /b 1
)
SET findCmd=FINDSTR /B "version=" "%dirName%\mageplusw"
FOR /f %%i IN ('%findCmd%') DO SET versionLine=%%i
SET version=%versionLine:~9,-1%

SET binariesCacheDir=%LOCALAPPDATA%\mageplus\binaries
IF NOT EXIST "%binariesCacheDir%" (
    md "%binariesCacheDir%"
)
IF NOT ERRORLEVEL 0 (
    CALL :fatal "Cannot create cache directory for storing binaries. See above."
)
SET binaryFileName=mageplus-%os%-%arch%-%version%%ext%
SET binary=%binariesCacheDir%\%binaryFileName%

IF NOT EXIST "%binary%" (
    CALL :doDownload
) ELSE (
    "%binary%" version 2>&1 | find "%version%" > NUL
    IF NOT ERRORLEVEL 0 (
        CALL :doDownload
    )
)

IF "%ERRORLEVEL%" == "0" (
    "%binary%" %*
)
EXIT /b %ERRORLEVEL%
GOTO :eofSuccess

:doDownload
    SETLOCAL
    SET binaryDownloadUrl=https://github.com/echocat/mageplus/releases/download/v%version%/mageplus_%version%_%os%-%arch%%downloadExt%
    CALL :info Downloading %binaryDownloadUrl%...

    SET tmpDirectory=%binary%.%RANDOM%.tmp
    PowerShell -Command "New-Item -Path '%tmpDirectory%' -Type Directory -Force | Out-Null; (New-Object Net.WebClient).DownloadFile('%binaryDownloadUrl%','%tmpDirectory%\tmp.zip'); Expand-Archive '%tmpDirectory%\tmp.zip' -DestinationPath '%tmpDirectory%'; Move-Item '%tmpDirectory%\mageplus.exe' '%binary%'; Remove-Item '%tmpDirectory%' -Recurse -Force"
    IF "%ERRORLEVEL%" NEQ "0" (
        CALL :fatal Was not able to download binary from %binaryDownloadUrl%. See above.
    )
    ENDLOCAL
    EXIT /b %ERRORLEVEL%

:fatal
    ECHO.FATAL: %*
    GOTO :eofError
    EXIT /b 1

:info
    ECHO.INFO: %*
    EXIT /b 0

:eofError
EXIT /b 1
GOTO :eof

:eofSuccess
EXIT /b 0
