document.addEventListener("click", (event) => {
  const trigger = event.target.closest("[data-copy-link]");
  if (!trigger) {
    return;
  }

  event.preventDefault();
  const value = trigger.getAttribute("data-copy-link");
  if (!value || !navigator.clipboard) {
    return;
  }

  navigator.clipboard.writeText(value).then(() => {
    trigger.textContent = "已复制";
    window.setTimeout(() => {
      trigger.textContent = "复制链接";
    }, 1200);
  });
});
