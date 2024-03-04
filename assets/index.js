(() => {
  const lastRead = document.querySelector("tr.read");
  const topForm = document.querySelector("form.log-button");
  const jumpBtn = document.querySelector(".jump-to-unread");

  if (!jumpBtn) {
    return;
  }

  if (topForm.classList.contains("catched-up")) {
    return;
  }

  if (!lastRead) {
    jumpBtn.classList.add("hidden");
    return;
  }

  jumpBtn.onclick = (evt) => {
    evt.preventDefault();
    lastRead.scrollIntoView(false);
  };

  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.target == topForm) {
        if (entry.isIntersecting) {
          jumpBtn.classList.remove("hidden");
        } else {
          jumpBtn.classList.add("hidden");
        }
      }

      if (entry.target == lastRead && entry.isIntersecting) {
        jumpBtn.classList.add("hidden");
      }
    }
  });
  observer.observe(topForm);
  observer.observe(lastRead);
})();
