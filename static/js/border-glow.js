const GRADIENT_POSITIONS = [
  "80% 55%",
  "69% 34%",
  "8% 6%",
  "41% 38%",
  "86% 85%",
  "82% 18%",
  "51% 4%",
];
const GRADIENT_KEYS = [
  "--gradient-one",
  "--gradient-two",
  "--gradient-three",
  "--gradient-four",
  "--gradient-five",
  "--gradient-six",
  "--gradient-seven",
];
const COLOR_MAP = [0, 1, 2, 0, 1, 2, 1];

function parseHSL(hslStr) {
  const match = hslStr.match(/([\d.]+)\s*([\d.]+)%?\s*([\d.]+)%?/);
  if (!match) return { h: 40, s: 80, l: 80 };
  return { h: parseFloat(match[1]), s: parseFloat(match[2]), l: parseFloat(match[3]) };
}

function buildGlowVars(glowColor, intensity) {
  const { h, s, l } = parseHSL(glowColor);
  const base = `${h}deg ${s}% ${l}%`;
  const opacities = [100, 60, 50, 40, 30, 20, 10];
  const keys = ["", "-60", "-50", "-40", "-30", "-20", "-10"];
  const vars = {};
  for (let i = 0; i < opacities.length; i++) {
    vars[`--glow-color${keys[i]}`] = `hsl(${base} / ${Math.min(opacities[i] * intensity, 100)}%)`;
  }
  return vars;
}

function buildGradientVars(colors) {
  const vars = {};
  for (let i = 0; i < 7; i++) {
    const c = colors[Math.min(COLOR_MAP[i], colors.length - 1)];
    vars[GRADIENT_KEYS[i]] = `radial-gradient(at ${GRADIENT_POSITIONS[i]}, ${c} 0px, transparent 50%)`;
  }
  vars["--gradient-base"] = `linear-gradient(${colors[0]} 0 100%)`;
  return vars;
}

function easeOutCubic(x) {
  return 1 - Math.pow(1 - x, 3);
}
function easeInCubic(x) {
  return x * x * x;
}

function animateValue({ start = 0, end = 100, duration = 1000, delay = 0, ease = easeOutCubic, onUpdate, onEnd }) {
  const t0 = performance.now() + delay;
  function tick() {
    const elapsed = performance.now() - t0;
    const t = Math.min(elapsed / duration, 1);
    onUpdate(start + (end - start) * ease(t));
    if (t < 1) requestAnimationFrame(tick);
    else if (onEnd) onEnd();
  }
  setTimeout(() => requestAnimationFrame(tick), delay);
}

function getCenterOfElement(el) {
  const { width, height } = el.getBoundingClientRect();
  return [width / 2, height / 2];
}

function getEdgeProximity(el, x, y) {
  const [cx, cy] = getCenterOfElement(el);
  const dx = x - cx;
  const dy = y - cy;
  let kx = Infinity;
  let ky = Infinity;
  if (dx !== 0) kx = cx / Math.abs(dx);
  if (dy !== 0) ky = cy / Math.abs(dy);
  return Math.min(Math.max(1 / Math.min(kx, ky), 0), 1);
}

function getCursorAngle(el, x, y) {
  const [cx, cy] = getCenterOfElement(el);
  const dx = x - cx;
  const dy = y - cy;
  if (dx === 0 && dy === 0) return 0;
  const radians = Math.atan2(dy, dx);
  let degrees = radians * (180 / Math.PI) + 90;
  if (degrees < 0) degrees += 360;
  return degrees;
}

function applyCardStyle(card, options) {
  const {
    glowColor = "40 80 80",
    backgroundColor = "#060010",
    borderRadius = 28,
    glowRadius = 40,
    glowIntensity = 1.0,
    coneSpread = 25,
    colors = ["#c084fc", "#f472b6", "#38bdf8"],
    fillOpacity = 0.5,
    edgeSensitivity = 30,
  } = options;

  card.style.setProperty("--card-bg", backgroundColor);
  card.style.setProperty("--edge-sensitivity", String(edgeSensitivity));
  card.style.setProperty("--border-radius", `${borderRadius}px`);
  card.style.setProperty("--glow-padding", `${glowRadius}px`);
  card.style.setProperty("--cone-spread", String(coneSpread));
  card.style.setProperty("--fill-opacity", String(fillOpacity));

  const glowVars = buildGlowVars(glowColor, glowIntensity);
  Object.entries(glowVars).forEach(([k, v]) => card.style.setProperty(k, v));
  const gradVars = buildGradientVars(colors);
  Object.entries(gradVars).forEach(([k, v]) => card.style.setProperty(k, v));
}

function bindPointer(card) {
  function onPointerMove(e) {
    const rect = card.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const edge = getEdgeProximity(card, x, y);
    const angle = getCursorAngle(card, x, y);
    card.style.setProperty("--edge-proximity", `${(edge * 100).toFixed(3)}`);
    card.style.setProperty("--cursor-angle", `${angle.toFixed(3)}deg`);
  }
  card.addEventListener("pointermove", onPointerMove);
  card.addEventListener("pointerleave", () => {
    card.style.setProperty("--edge-proximity", "0");
  });
  return () => {
    card.removeEventListener("pointermove", onPointerMove);
  };
}

function runSweepAnimation(card) {
  const angleStart = 110;
  const angleEnd = 465;
  card.classList.add("sweep-active");
  card.style.setProperty("--cursor-angle", `${angleStart}deg`);

  animateValue({ duration: 500, onUpdate: (v) => card.style.setProperty("--edge-proximity", String(v)) });
  animateValue({
    ease: easeInCubic,
    duration: 1500,
    end: 50,
    onUpdate: (v) => {
      card.style.setProperty("--cursor-angle", `${(angleEnd - angleStart) * (v / 100) + angleStart}deg`);
    },
  });
  animateValue({
    ease: easeOutCubic,
    delay: 1500,
    duration: 2250,
    start: 50,
    end: 100,
    onUpdate: (v) => {
      card.style.setProperty("--cursor-angle", `${(angleEnd - angleStart) * (v / 100) + angleStart}deg`);
    },
  });
  animateValue({
    ease: easeInCubic,
    delay: 2500,
    duration: 1500,
    start: 100,
    end: 0,
    onUpdate: (v) => card.style.setProperty("--edge-proximity", String(v)),
    onEnd: () => card.classList.remove("sweep-active"),
  });
}

/**
 * @param {HTMLElement} root
 */
export function initBorderGlowCards(root = document) {
  const cards = root.querySelectorAll(".border-glow-card");
  const cleanups = [];
  cards.forEach((card) => {
    const options = {
      glowColor: card.dataset.glowColor || "200 70 70",
      backgroundColor: card.dataset.cardBg || "#0a0614",
      borderRadius: Number(card.dataset.borderRadius || 24),
      glowRadius: Number(card.dataset.glowRadius || 36),
      glowIntensity: Number(card.dataset.glowIntensity || 1),
      coneSpread: Number(card.dataset.coneSpread || 25),
      colors: (card.dataset.colors || "").split("|").filter(Boolean),
      fillOpacity: Number(card.dataset.fillOpacity || 0.45),
      edgeSensitivity: Number(card.dataset.edgeSensitivity || 30),
    };
    if (options.colors.length === 0) {
      options.colors = ["#a78bfa", "#f472b6", "#38bdf8"];
    }
    applyCardStyle(card, options);
    cleanups.push(bindPointer(card));
    if (card.dataset.animated === "true") {
      runSweepAnimation(card);
    }
  });
  return () => cleanups.forEach((fn) => fn());
}
