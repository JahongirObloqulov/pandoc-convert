!define APP_NAME "Pandoc Convert Pro"
!define COMP_NAME "JahongirObloqulov"
!define VERSION "1.0.0"
!define OUT_FILE "build\bin\pandoc-pro-installer.exe"
!define REG_UNINSTALL "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

InstallDir "$PROGRAMFILES\PandocConvertPro"

Name "${APP_NAME}"
OutFile "${OUT_FILE}"
RequestExecutionLevel admin

Page Directory
Page InstFiles

Section "Install"
    SetOutPath "$INSTDIR"
    
    ; Copy files
    File "build\bin\pandoc-app.exe"
    File "C:\Users\Cyber\AppData\Local\Pandoc\pandoc.exe"
    
    ; Create shortcuts
    CreateShortcut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\pandoc-app.exe"
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\pandoc-app.exe"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\Uninstall.lnk" "$INSTDIR\uninstall.exe"
    
    ; Write uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"
    
    ; Write registry for Add/Remove Programs
    WriteRegStr HKLM "${REG_UNINSTALL}" "DisplayName" "${APP_NAME}"
    WriteRegStr HKLM "${REG_UNINSTALL}" "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKLM "${REG_UNINSTALL}" "DisplayIcon" "$INSTDIR\pandoc-app.exe"
    WriteRegStr HKLM "${REG_UNINSTALL}" "DisplayVersion" "${VERSION}"
    WriteRegStr HKLM "${REG_UNINSTALL}" "Publisher" "${COMP_NAME}"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\pandoc-app.exe"
    Delete "$INSTDIR\pandoc.exe"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"
    
    Delete "$DESKTOP\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\Uninstall.lnk"
    RMDir "$SMPROGRAMS\${APP_NAME}"
    
    ; Delete registry
    DeleteRegKey HKLM "${REG_UNINSTALL}"
SectionEnd
