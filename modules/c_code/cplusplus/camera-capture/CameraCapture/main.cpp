#define _UNICODE
#include <tchar.h>
#include <windows.h>
#include <vfw.h>

LRESULT capVideoCallBack(HWND hWnd, LPVIDEOHDR lpVHdr)
{
    WCHAR tempPath[MAX_PATH];
    GetTempPathW(MAX_PATH, tempPath);

    WCHAR bmpFilePath[MAX_PATH];
    wcscpy_s(bmpFilePath, tempPath);
    wcscat_s(bmpFilePath, L"Image.bmp");

    HANDLE fileHandle = CreateFileW(bmpFilePath, GENERIC_WRITE, 0, NULL, CREATE_ALWAYS, FILE_ATTRIBUTE_NORMAL, NULL);
    WriteFile(fileHandle, lpVHdr->lpData, lpVHdr->dwBytesUsed, nullptr, NULL);
    CloseHandle(fileHandle);

    // Set a timer to delete the AVI file after a delay.
    SetTimer(hWnd, 1, 2000, NULL); // 2000 milliseconds (2 seconds)

    return 0;
}

VOID CALLBACK TimerProc(HWND hWnd, UINT uMsg, UINT_PTR idEvent, DWORD dwTime)
{
    WCHAR tempPath[MAX_PATH];
    GetTempPathW(MAX_PATH, tempPath);

    WCHAR aviFilePath[MAX_PATH];
    wcscpy_s(aviFilePath, tempPath);
    wcscat_s(aviFilePath, L"Image.avi");

    DeleteFileW(aviFilePath);

    KillTimer(hWnd, 1);
}

INT wWinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, LPWSTR lpCmdLine, INT nShowCmd)
{
    // Create a capture window.
    HWND hWnd = capCreateCaptureWindowW(L" ", 0, 0, 0, 0, 0, 0, 0);
    capSetCallbackOnFrame(hWnd, capVideoCallBack);
    // Connect to a device and then capture a single frame from the webcam while writing it to a dummy AVI file which can be discarded.
    capDriverConnect(hWnd, 0);
    capFileSetCaptureFile(hWnd, _T("Image.avi"));
    capCaptureSingleFrameOpen(hWnd);
    capCaptureSingleFrame(hWnd);
    capCaptureSingleFrameClose(hWnd);
    capFileSaveAs(hWnd, _T("Image.avi"));
    capDriverDisconnect(hWnd);
    DestroyWindow(hWnd);

    MSG msg;
    while (GetMessage(&msg, NULL, 0, 0))
    {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }

    return EXIT_SUCCESS;
}
