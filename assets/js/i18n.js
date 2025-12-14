window.toggleLangMenu = function () {
  const menu = document.getElementById("lang-menu");
  if (!menu) return;
  menu.classList.toggle("hidden");
};

// click outside to close dropdown
document.addEventListener("click", (e) => {
  const toggle = document.getElementById("lang-toggle");
  const menu = document.getElementById("lang-menu");
  if (!toggle || !menu) return;

  if (!toggle.contains(e.target) && !menu.contains(e.target)) {
    menu.classList.add("hidden");
  }
});
