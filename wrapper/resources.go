package wrapper

func init() {
	unixScript = `#!/bin/sh
##############################################################################
##                                                                          ##
##  mageplus bootstrap wrapper for *NIX systems                             ##
##                                                                          ##
##############################################################################
version="####VERSION####"

##############################################################################
##  DO NOT EDIT BELOW!!                                                     ##
##############################################################################
fatal() {
    echo "FATAL: $*" 1>&2
    exit 1
}

info() {
    echo "INFO: $*" 1>&2
}

doDownload() {
    binaryDownloadUrl="https://github.com/echocat/mageplus/releases/download/v${version}/mageplus_${version}_${os}-${arch}${downloadExt}"
    tmpDirectory="${binary}.tmp"
    info "Downloading ${binaryDownloadUrl}..."

    mkdir -p "${tmpDirectory}"
    if [ "$?" != "0" ]; then
        fatal "Cannot create cache directory for storing binaries. See above."
    fi

    if command -v curl > /dev/null; then
        curl -sSLf "${binaryDownloadUrl}" > "${tmpDirectory}/tmp.tar.gz"
        if [ "$?" != "0" ]; then
            fatal "Was not able to download binary from ${binaryDownloadUrl}. See above."
        fi
    elif command -v wget > /dev/null; then
        wget -q -O "${tmpDirectory}/tmp.tar.gz" "${binaryDownloadUrl}"
        if [ "$?" != "0" ]; then
            fatal "Was not able to download binary from ${binaryDownloadUrl}. See above."
        fi
    else
        fatal "Neither curl nor wget found in \$PATH. Please install at least one of those tools."
    fi

    tar -xzf "${tmpDirectory}/tmp.tar.gz" -C "${tmpDirectory}"
    if [ "$?" != "0" ]; then
        fatal "Was not able to extract ${tmpDirectory}/tmp.tar.gz. See above."
    fi

    chmod +x "${tmpDirectory}/mageplus"
    if [ "$?" != "0" ]; then
        fatal "Was not able to make ${tmpDirectory}/mageplus executable. See above."
    fi

    mv "${tmpDirectory}/mageplus" "${binary}"
    if [ "$?" != "0" ]; then
        fatal "Was not able to move ${tmpDirectory}/mageplus to ${binary}. See above."
    fi

    rm -rf "${tmpDirectory}"
    if [ "$?" != "0" ]; then
        fatal "Was not able to clean up ${tmpDirectory}. See above."
    fi
}

plainOs="$(uname -s)"
case "${plainOs}" in
    Linux*)        os="Linux";;
    Darwin*)       os="macOS";;
    FreeBSD*)      os="FreeBSD";;
    OpenBSD*)      os="OpenBSD";;
    NetBSD*)       os="NetBSD";;
    DragonFlyBSD*) os="DragonFlyBSD";;
    CYGWIN*)       os="Windows";;
    MINGW*)        os="Windows";;
    *)             fatal "Unsupported operating system: ${plainOs}"
esac

plainArch="$(uname -m)"
case "${plainArch}" in
    x86_64*)       arch="64bit";;
    i386*)         arch="32bit";;
    arm64*)        arch="ARM64";;
    arm*)          arch="ARM";;
    *)             fatal "Unsupported architecture: ${plainArch}"
esac

case "${os}" in
    windows*)   ext=".exe";;
    *)          ext="";;
esac

case "${os}" in
    windows*)   downloadExt=".zip";;
    *)          downloadExt=".tar.gz";;
esac

binariesCacheDir="${HOME}/.mageplus/binaries"
binaryFileName="mageplus-${os}-${arch}-${version}${ext}"
binary="${binariesCacheDir}/${binaryFileName}"

if [ "${MAGEPLUSW_IGNORE_DOCKER_IMAGE_MISMATCH}" != "yes" ]; then
    if [ -r "/usr/lib/mageplus/docker-version" ]; then
        dockerVersion="$(cat /usr/lib/mageplus/docker-version)"
        if [ "${dockerVersion}" != "${version}" ]; then
            if [ -r "/usr/lib/mageplus/docker-image" ]; then
                dockerImage="$(cat /usr/lib/mageplus/docker-image)"
            else
                dockerImage="echocat/mageplus"
            fi
            fatal "You're are using mageplusw with version ${version} inside of a mageplus docker image with version ${dockerVersion}." \
                  "This could lead to unexpected behaviors. We recommend to align both versions together by either:" \
                  "\n\t1.) Change ${0} to: version=\"${dockerVersion}\"" \
                  "\n\t2.) ... or set the used image to: ${dockerImage}:${version}" \
                  "\nYou can suppress this error by set MAGEPLUSW_IGNORE_DOCKER_IMAGE_MISMATCH=yes"
        fi
    fi
fi

if [ -x "${binary}" ]; then
    "${binary}" --version 2>&1 | grep "${version}" > /dev/null
    if [ "$?" != "0" ]; then
        doDownload
    fi
else
    doDownload
fi

"${binary}" "$@"
`
	windowsScript = `@ECHO OFF
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
`
}
