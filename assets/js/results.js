// confetti on page load
window.addEventListener("load", () => {
  const end = Date.now() + 2000;
  (function frame() {
    confetti({ particleCount: 5, spread: 70, origin: { y: 0.6 } });
    if (Date.now() < end) requestAnimationFrame(frame);
  })();
});
