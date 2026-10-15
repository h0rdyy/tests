import { mountDarkVeil } from "./darkveil.js";
import { initBorderGlowCards } from "./border-glow.js";
import { initScrollHeader, initLeadModal, initLeadForms } from "./site-common.js";
import { initCasesCarousel } from "./cases-carousel.js";


const canvas = document.getElementById("darkveil-canvas");
if (canvas) {
  mountDarkVeil(canvas, {
    noiseIntensity: 0.02,
    warpAmount: 0.08,
    speed: 0.45,
  });
}


initScrollHeader(document.getElementById("site-header"));
initLeadModal();
initLeadForms();
initBorderGlowCards(document);

const casesRoot = document.getElementById("cases-root");
if (casesRoot) initCasesCarousel(casesRoot);


document.addEventListener("DOMContentLoaded", () => {
  // Находим ВСЕ формы заявок на странице
  const allLeadForms = document.querySelectorAll(".lead-form");

  allLeadForms.forEach((form) => {
    const formMsg = form.querySelector("[data-form-msg]");     // Нижний блок (под кнопкой)
    const fieldMsg = form.querySelector("[data-field-msg]");   // Блок под номером телефона
    const phoneInput = form.querySelector('input[name="phone"]'); // Само поле телефона

    // ПЕРЕНОС ОШИБКИ ТЕЛЕФОНА НАВЕРХ (ЧЕРЕЗ OBSERVER)
    if (formMsg && fieldMsg && phoneInput) {
      const observer = new MutationObserver(() => {
        const text = formMsg.textContent.trim();
        
        // Если пришла ошибка про телефон (содержит слово "телефон" или "корректный")
        if (text && (text.toLowerCase().includes("телефон") || text.toLowerCase().includes("корректный"))) {
          fieldMsg.textContent = text;
          fieldMsg.hidden = false; // Показываем под полем
          phoneInput.style.border = "1px solid #ff4d4f"; // Красим рамку телефона в красный
          
          // Очищаем нижнее сообщение, чтобы оно не дублировалось
          formMsg.textContent = ""; 
          formMsg.className = "lead-form__msg"; // Сбрасываем классы
        } 
      });

      // Запускаем слежку за нижним блоком
      observer.observe(formMsg, { childList: true, characterData: true, subtree: true });

      // Когда пользователь начинает заново вводить номер - прячем красную ошибку
      phoneInput.addEventListener("input", () => {
        fieldMsg.hidden = true;
        fieldMsg.textContent = "";
        phoneInput.style.border = ""; // Возвращаем обычную рамку
      });
    }
  });
});