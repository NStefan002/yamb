// constants for popup behavior
const BASE_BOTTOM_OFFSET = 120; // distance from footer to first popup
const STACK_SPACING = 60; // vertical distance between stacked popups
const DISMISS_SLIDE = 40; // px to slide out when dismissed
const AUTO_DISMISS_MS = 7000; // per-popup timer
const FADE_IN_DELAY_MS = 1500; // time before start glide animation

// state
let activeErrors = [];

// show new popup
function showFloatingError(msg) {
  const container = document.getElementById("error-container");
  const el = document.createElement("div");

  const color = "#dc2626";

  el.className = `
    text-white font-semibold px-5 py-3 rounded-lg shadow-xl cursor-pointer
    fixed z-[99999] pointer-events-auto select-none
    transition-all duration-500 ease-out
    opacity-0 scale-90
  `;
  el.style.backgroundColor = color;
  el.textContent = msg;

  // click → dismiss immediately (with animation)
  el.addEventListener("click", () => dismissError(el));

  container.appendChild(el);
  activeErrors.push(el);

  // 1) spawn centered, fade/scale in
  el.style.top = "50%";
  el.style.left = "50%";
  el.style.transform = "translate(-50%, -50%) scale(0.9)";
  el.style.opacity = "0";

  requestAnimationFrame(() => {
    el.style.opacity = "1";
    el.style.transform = "translate(-50%, -50%) scale(1)";
  });

  // 2) after fade-in, glide to dock spot
  setTimeout(() => animateToDock(el), FADE_IN_DELAY_MS);

  // 3) auto-dismiss after timeout (each popup has its own timer)
  const timerId = setTimeout(() => dismissError(el), AUTO_DISMISS_MS);
  el.dataset.timerId = timerId;
}

// animate popup from center → stacked bottom-right position
function animateToDock(el) {
  el.getBoundingClientRect();

  const index = activeErrors.indexOf(el);
  const offset = index * STACK_SPACING;

  const screenW = window.innerWidth;
  const screenH = window.innerHeight;

  const popupWidth = el.offsetWidth;

  const bottomOffset = BASE_BOTTOM_OFFSET + offset;
  const rightOffset = 20;

  const finalX = screenW / 2 - rightOffset - popupWidth;
  const finalY = screenH / 2 - bottomOffset;

  el.style.transform = `translate(${finalX}px, ${finalY}px) scale(0.95)`;
  el.style.opacity = "0.95";
}

// remove popup (click or timer) with smooth exit animation
function dismissError(el) {
  if (el.dataset.timerId) clearTimeout(el.dataset.timerId);

  el.style.transform += ` translateX(${DISMISS_SLIDE}px) scale(0.9)`;
  el.style.opacity = "0";

  setTimeout(() => {
    activeErrors = activeErrors.filter((e) => e !== el);
    el.remove();
    repositionStack();
  }, 500); // matches CSS transition
}

// after one popup is removed, move remaining popups upward to close gap
function repositionStack() {
  activeErrors.forEach((el, index) => {
    const offset = index * STACK_SPACING;

    const screenW = window.innerWidth;
    const screenH = window.innerHeight;

    const popupWidth = el.offsetWidth;

    const bottomOffset = BASE_BOTTOM_OFFSET + offset;
    const rightOffset = 20;

    const finalX = screenW / 2 - rightOffset - popupWidth;
    const finalY = screenH / 2 - bottomOffset;

    el.style.transform = `translate(${finalX}px, ${finalY}px) scale(0.95)`;
  });
}

// HTMX hook - listen for HX-Trigger: {"showError": "..."}
// [look at the handler.go]
document.addEventListener("htmx:afterOnLoad", (e) => {
  const trigger = e.detail.xhr.getResponseHeader("HX-Trigger");
  if (!trigger) return;

  try {
    const data = JSON.parse(trigger);
    if (data.showError) showFloatingError(data.showError);
  } catch {}
});
