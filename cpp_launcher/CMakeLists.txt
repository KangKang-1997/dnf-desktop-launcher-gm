cmake_minimum_required(VERSION 3.24)

project(dnf_webview2_launcher LANGUAGES CXX RC)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

set(APP_ICON_PATH "${CMAKE_CURRENT_SOURCE_DIR}/assets/app.ico")
set(INTERCEPTION_INSTALLER_PATH "${CMAKE_CURRENT_SOURCE_DIR}/assets/interception/install-interception.exe")
set(INTERCEPTION_DLL_X64_PATH "${CMAKE_CURRENT_SOURCE_DIR}/assets/interception/x64/interception.dll")
set(FRONTEND_HTML_PATH "${CMAKE_CURRENT_BINARY_DIR}/embedded-frontend.html" CACHE FILEPATH "Path to embedded frontend HTML")
set(DEFAULT_BACKGROUND_PATH "${CMAKE_CURRENT_SOURCE_DIR}/../desktop_launcher/public/assets/background.jpg" CACHE FILEPATH "Path to embedded launcher background image")
if(NOT EXISTS "${INTERCEPTION_INSTALLER_PATH}")
    message(FATAL_ERROR "Interception installer not found: ${INTERCEPTION_INSTALLER_PATH}")
endif()
if(NOT EXISTS "${INTERCEPTION_DLL_X64_PATH}")
    message(FATAL_ERROR "Interception x64 DLL not found: ${INTERCEPTION_DLL_X64_PATH}")
endif()
if(NOT EXISTS "${FRONTEND_HTML_PATH}")
    message(FATAL_ERROR "Embedded frontend HTML not found: ${FRONTEND_HTML_PATH}")
endif()
if(NOT EXISTS "${DEFAULT_BACKGROUND_PATH}")
    message(FATAL_ERROR "Default background image not found: ${DEFAULT_BACKGROUND_PATH}")
endif()
configure_file(
    "${CMAKE_CURRENT_SOURCE_DIR}/src/app.rc.in"
    "${CMAKE_CURRENT_BINARY_DIR}/app.rc"
    @ONLY
)

add_executable(dnf-webview2-launcher WIN32
    src/main.cpp
    "${CMAKE_CURRENT_BINARY_DIR}/app.rc"
)

target_compile_definitions(dnf-webview2-launcher PRIVATE
    UNICODE
    _UNICODE
    NOMINMAX
)

if(MSVC)
    target_compile_options(dnf-webview2-launcher PRIVATE /utf-8 /W4 /permissive-)
endif()

set(WEBVIEW2_SDK_DIR "$ENV{WEBVIEW2_SDK_DIR}" CACHE PATH "Path to Microsoft.Web.WebView2 native SDK directory")

if(NOT WEBVIEW2_SDK_DIR)
    if(DEFINED ENV{USERPROFILE})
        file(GLOB WEBVIEW2_PACKAGE_DIRS LIST_DIRECTORIES true "$ENV{USERPROFILE}/.nuget/packages/microsoft.web.webview2/*")
        list(SORT WEBVIEW2_PACKAGE_DIRS COMPARE NATURAL ORDER DESCENDING)
        list(LENGTH WEBVIEW2_PACKAGE_DIRS WEBVIEW2_PACKAGE_DIR_COUNT)
        if(WEBVIEW2_PACKAGE_DIR_COUNT GREATER 0)
            list(GET WEBVIEW2_PACKAGE_DIRS 0 WEBVIEW2_PACKAGE_DIR)
            set(WEBVIEW2_SDK_DIR "${WEBVIEW2_PACKAGE_DIR}/build/native")
        endif()
    endif()
endif()

if(NOT WEBVIEW2_SDK_DIR OR NOT EXISTS "${WEBVIEW2_SDK_DIR}/include/WebView2.h")
    message(FATAL_ERROR "WebView2 SDK not found. Run: dotnet restore cpp_launcher/WebView2SdkRestore.csproj")
endif()

target_include_directories(dnf-webview2-launcher PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/src"
    "${WEBVIEW2_SDK_DIR}/include"
)

if(CMAKE_SIZEOF_VOID_P EQUAL 8)
    set(WEBVIEW2_ARCH "x64")
else()
    set(WEBVIEW2_ARCH "x86")
endif()

target_link_libraries(dnf-webview2-launcher PRIVATE
    "${WEBVIEW2_SDK_DIR}/${WEBVIEW2_ARCH}/WebView2LoaderStatic.lib"
    Shlwapi
    Crypt32
    Shell32
)
