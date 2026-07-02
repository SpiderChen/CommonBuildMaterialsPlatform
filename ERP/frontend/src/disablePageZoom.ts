const zoomShortcutKeys = new Set(["+", "=", "-", "_", "0"]);
const zoomShortcutCodes = new Set(["Equal", "Minus", "Digit0", "NumpadAdd", "NumpadSubtract", "Numpad0"]);
const gestureEvents = ["gesturestart", "gesturechange", "gestureend"];

type ZoomGuardWindow = Window & {
  __cbmpPageZoomDisabled?: boolean;
  __cbmpNativeContextMenuDisabled?: boolean;
};

export function disablePageZoom() {
  const zoomWindow = window as ZoomGuardWindow;
  if (zoomWindow.__cbmpPageZoomDisabled) {
    return;
  }
  zoomWindow.__cbmpPageZoomDisabled = true;

  window.addEventListener(
    "wheel",
    (event) => {
      if (event.ctrlKey || event.metaKey) {
        event.preventDefault();
      }
    },
    { passive: false }
  );

  window.addEventListener("keydown", (event) => {
    if ((event.ctrlKey || event.metaKey) && (zoomShortcutKeys.has(event.key) || zoomShortcutCodes.has(event.code))) {
      event.preventDefault();
    }
  });

  gestureEvents.forEach((eventName) => {
    document.addEventListener(
      eventName,
      (event) => {
        event.preventDefault();
      },
      { passive: false }
    );
  });
}

export function disableNativeContextMenu() {
  const guardedWindow = window as ZoomGuardWindow;
  if (guardedWindow.__cbmpNativeContextMenuDisabled) {
    return;
  }
  guardedWindow.__cbmpNativeContextMenuDisabled = true;

  window.addEventListener(
    "contextmenu",
    (event) => {
      event.preventDefault();
    },
    { capture: true }
  );
}
