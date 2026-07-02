const zoomShortcutKeys = new Set(["+", "=", "-", "_", "0"]);
const zoomShortcutCodes = new Set(["Equal", "Minus", "Digit0", "NumpadAdd", "NumpadSubtract", "Numpad0"]);
const gestureEvents = ["gesturestart", "gesturechange", "gestureend"];

export function disablePageZoom() {
  if (window.__cbmpPageZoomDisabled) {
    return;
  }
  window.__cbmpPageZoomDisabled = true;

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
  if (window.__cbmpNativeContextMenuDisabled) {
    return;
  }
  window.__cbmpNativeContextMenuDisabled = true;

  window.addEventListener(
    "contextmenu",
    (event) => {
      event.preventDefault();
    },
    { capture: true }
  );
}
