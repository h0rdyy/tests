const API_LEAD = "/api/lead";
const FIELD_REQUIRED_MSG = "Важно заполнить это поле";

export function initScrollHeader(headerEl) {
  if (!headerEl) return () => {};

  let lastY = window.scrollY;
  const threshold = 8;

  function onScroll() {
    const y = window.scrollY;
    const dy = y - lastY;

    if (y <= threshold) {
      headerEl.classList.remove("header--hidden", "header--glass");
    } else if (dy > 2) {
      headerEl.classList.add("header--hidden");
      headerEl.classList.remove("header--glass");
    } else if (dy < -2) {
      headerEl.classList.remove("header--hidden");
      headerEl.classList.add("header--glass");
    }

    lastY = y;
  }

  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();

  return () => window.removeEventListener("scroll", onScroll);
}

export function initLeadModal() {
  const backdrop = document.getElementById("modal-backdrop");
  const modal = document.getElementById("lead-modal");
  const openBtn = document.getElementById("open-lead-modal");
  const closeBtn = document.getElementById("close-lead-modal");

  function open() {
    backdrop?.classList.add("is-open");
    modal?.classList.add("is-open");
    backdrop?.setAttribute("aria-hidden", "false");
    modal?.setAttribute("aria-hidden", "false");
    document.body.style.overflow = "hidden";
  }

  function close() {
    backdrop?.classList.remove("is-open");
    modal?.classList.remove("is-open");
    backdrop?.setAttribute("aria-hidden", "true");
    modal?.setAttribute("aria-hidden", "true");
    document.body.style.overflow = "";
  }

  openBtn?.addEventListener("click", open);
  closeBtn?.addEventListener("click", close);
  backdrop?.addEventListener("click", close);
  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape") close();
  });

  return { open, close };
}

function setFormMessage(form, text, kind) {
  const msg = form.querySelector("[data-form-msg]");
  if (!msg) return;
  msg.textContent = text || "";
  msg.classList.remove("lead-form__msg--ok", "lead-form__msg--err");
  if (kind === "ok") msg.classList.add("lead-form__msg--ok");
  if (kind === "err") msg.classList.add("lead-form__msg--err");
}

function setFieldMessage(form, text) {
  const el = form.querySelector("[data-field-msg]");
  if (!el) return;
  const show = Boolean(text);
  el.textContent = text || "";
  el.hidden = !show;
  const input = form.querySelector('input[name="phone"]');
  if (input) input.setAttribute("aria-invalid", show ? "true" : "false");
}

function clearFieldMessage(form) {
  setFieldMessage(form, "");
}

export function initLeadForms() {
  const forms = document.querySelectorAll("form.lead-form[data-source]");
  const controllers = [];

  forms.forEach((form) => {
    const phoneInput = form.querySelector('input[name="phone"]');

    function onInput() {
      clearFieldMessage(form);
      setFormMessage(form, "", null);
    }

    phoneInput?.addEventListener("input", onInput);

    async function onSubmit(e) {
      e.preventDefault();
      const fd = new FormData(form);
      const phone = (fd.get("phone") || "").trim();
      const source = form.getAttribute("data-source") || "form";

      if (!phone) {
        setFieldMessage(form, FIELD_REQUIRED_MSG);
        setFormMessage(form, "", null);
        phoneInput?.focus();
        return;
      }

      clearFieldMessage(form);
      setFormMessage(form, "Отправка…", null);

      try {
        const res = await fetch(API_LEAD, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ phone, source }),
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          setFormMessage(form, data.error || "Не удалось отправить. Попробуйте позже.", "err");
          return;
        }
        setFormMessage(form, "Мы свяжемся с Вами!", "ok");
        form.reset();
      } catch {
        setFormMessage(form, "Ошибка сети. Проверьте подключение.", "err");
      }
    }

    form.addEventListener("submit", onSubmit);
    controllers.push(() => {
      form.removeEventListener("submit", onSubmit);
      phoneInput?.removeEventListener("input", onInput);
    });
  });

  return () => controllers.forEach((fn) => fn());
}
