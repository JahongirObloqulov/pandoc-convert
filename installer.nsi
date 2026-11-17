!define APP_NAME "Pandoc Convert Pro"
!define COMP_NAME "JahongirObloqulov"
!define VERSION "1.1.0"
!define OUT_FILE "build\bin\pandoc-pro-installer.exe"
!define REG_UNINSTALL "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_NAME}"

; 64-bitli ilova: "Program Files (x86)" emas, "Program Files" ga o'rnatiladi
InstallDir "$PROGRAMFILES64\PandocConvertPro"

Name "${APP_NAME}"
OutFile "${OUT_FILE}"
Icon "build\windows\icon.ico"
UninstallIcon "build\windows\icon.ico"
RequestExecutionLevel admin

Page Directory
Page InstFiles

Section "Install"
    ; Yorliqlar barcha foydalanuvchilar uchun (HKLM ro'yxatiga mos)
    SetShellVarContext all
    SetOutPath "$INSTDIR"
    
    ; Copy files
    File "build\bin\pandoc-app.exe"
    File "build\bin\pandoc.exe"
    File "build\windows\icon.ico"
    
    ; Create shortcuts
    CreateShortcut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\pandoc-app.exe" "" "$INSTDIR\icon.ico" 0
    CreateDirectory "$SMPROGRAMS\${APP_NAME}"
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk" "$INSTDIR\pandoc-app.exe" "" "$INSTDIR\icon.ico" 0
    CreateShortcut "$SMPROGRAMS\${APP_NAME}\Uninstall.lnk" "$INSTDIR\uninstall.exe"

    ; Refresh Windows icon cache so desktop shortcut shows the icon immediately
    System::Call 'Shell32::SHChangeNotify(i 0x08000000, i 0, i 0, i 0)'
    
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
    SetShellVarContext all
    Delete "$INSTDIR\pandoc-app.exe"
    Delete "$INSTDIR\pandoc.exe"
    Delete "$INSTDIR\icon.ico"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR"
    
    Delete "$DESKTOP\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\${APP_NAME}.lnk"
    Delete "$SMPROGRAMS\${APP_NAME}\Uninstall.lnk"
    RMDir "$SMPROGRAMS\${APP_NAME}"
    
    ; Delete registry
    DeleteRegKey HKLM "${REG_UNINSTALL}"
SectionEnd
