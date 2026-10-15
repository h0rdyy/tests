import * as THREE from "three";
import { createAsciiArt, coarseLowPower } from "../vendor/tre/core.js";

const VIOLET = "#8b5cf6";
const BOOT_SECONDS = 1.6;
const CAM_Z = 5;
const FOV = 45;

const clamp = (v, min, max) => Math.min(max, Math.max(min, v));

/**
 * Живой ASCII/CRT-фон с логотипом ЦККБ — настоящая 3D-сцена внутри движка
 * @borkiss/tre. Логотип — плоскость в пространстве, вокруг облако фиолетовой
 * «пыли» на разной глубине; камера плавно летает за курсором (перспектива и
 * параллакс слоёв), при наведении на панель приближается, при быстром рывке
 * мыши сигнал «срывается» (datamosh + тряска). Движок tre превращает всё это
 * в глифы с фосфорным послесвечением и «включением» при появлении на экране.
 *
 * @param {HTMLCanvasElement} canvas  должен иметь CSS-размер (absolute внутри карточки)
 * @param {{ logoSrc?: string, ghostScale?: number, ghostX?: number }} [opts]
 *        ghostScale — во сколько раз ASCII-«призрак» логотипа больше базового;
 *        ghostX — центр логотипа по горизонтали (0..1).
 * @returns {() => void} функция очистки
 */
export function mountAboutAscii(canvas, opts = {}) {
  if (!canvas) return () => {};

  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  const ghostScale = opts.ghostScale || 1;
  const ghostX = opts.ghostX;

  // — логотип: печём в 2D-канвас со свечением и контрастом → текстура —
  const logoCanvas = document.createElement("canvas");
  const lctx = logoCanvas.getContext("2d");
  let logoTexture = null;

  const img = new Image();
  img.decoding = "async";

  function paintLogo() {
    const iw = img.naturalWidth;
    const ih = img.naturalHeight;
    if (!iw || !ih) return;

    const W = 1024;
    const H = Math.max(256, Math.round((W * ih) / iw));
    logoCanvas.width = W;
    logoCanvas.height = H;
    const lw = W * 0.86;
    const lh = (lw * ih) / iw;
    const dx = (W - lw) / 2;
    const dy = (H - lh) / 2;

    lctx.clearRect(0, 0, W, H);
    // мягкое запечённое свечение вокруг логотипа…
    lctx.save();
    lctx.shadowColor = "rgba(124, 58, 237, 0.6)";
    lctx.shadowBlur = H * 0.07;
    lctx.drawImage(img, dx, dy, lw, lh);
    lctx.restore();
    // …и сам логотип с усиленным контрастом: неоновые грани становятся
    // яркими, тёмная поверхность стекла уходит в чёрный — глифы читаются.
    lctx.save();
    lctx.filter = "contrast(1.7) brightness(1.25) saturate(1.2)";
    lctx.drawImage(img, dx, dy, lw, lh);
    lctx.restore();

    const next = new THREE.CanvasTexture(logoCanvas);
    next.colorSpace = THREE.SRGBColorSpace;
    if (logoTexture) logoTexture.dispose();
    logoTexture = next;
    logoPlane.material.map = next;
    logoPlane.material.needsUpdate = true;
    logoPlane.visible = true;
    layout();
  }

  // — 3D-сцена: плоскость логотипа + пыль глубины —
  const scene = new THREE.Scene();
  scene.background = new THREE.Color("#05030d");

  const camera = new THREE.PerspectiveCamera(FOV, 1, 0.1, 50);
  camera.position.set(0, 0, CAM_Z);

  const logoPlane = new THREE.Mesh(
    new THREE.PlaneGeometry(1, 1),
    new THREE.MeshBasicMaterial({ transparent: true, depthWrite: false })
  );
  logoPlane.visible = false;
  scene.add(logoPlane);

  const DUST_COUNT = 240;
  const dustPos = new Float32Array(DUST_COUNT * 3);
  for (let i = 0; i < DUST_COUNT; i++) {
    dustPos[i * 3] = (Math.random() * 2 - 1) * 3.6;
    dustPos[i * 3 + 1] = (Math.random() * 2 - 1) * 2.3;
    dustPos[i * 3 + 2] = -2.2 + Math.random() * 3.0; // глубина вокруг логотипа
  }
  const dustGeo = new THREE.BufferGeometry();
  dustGeo.setAttribute("position", new THREE.BufferAttribute(dustPos, 3));
  const dust = new THREE.Points(
    dustGeo,
    new THREE.PointsMaterial({
      color: 0xa78bfa,
      size: 0.05,
      sizeAttenuation: true,
      transparent: true,
      opacity: 0.6,
      depthWrite: false,
      blending: THREE.AdditiveBlending,
    })
  );
  scene.add(dust);

  function layout() {
    const w = canvas.clientWidth;
    const h = canvas.clientHeight;
    if (!w || !h) return;
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
    // видимая область на глубине плоскости логотипа (z = 0)
    const visH = 2 * CAM_Z * Math.tan(THREE.MathUtils.degToRad(FOV / 2));
    const visW = visH * camera.aspect;
    const portrait = w / h < 0.95;
    const cx = ghostX !== undefined ? ghostX : portrait ? 0.5 : 0.72;
    const logoH = visH * (portrait ? 0.62 : 0.95) * ghostScale;
    const logoW = logoH * (logoCanvas.width / (logoCanvas.height || 1));
    logoPlane.scale.set(logoW || 1, logoH, 1);
    logoPlane.position.set((cx - 0.5) * visW, 0, 0);
    // облако пыли держится вокруг логотипа
    dust.position.x = logoPlane.position.x;
  }

  // — курсор: камера догоняет его, быстрый рывок даёт «срыв сигнала» —
  const pointer = { x: 0, y: 0, tx: 0, ty: 0, hover: 0, hovering: false, kick: 0 };

  const onPointerMove = (e) => {
    const r = canvas.getBoundingClientRect();
    if (!r.width || !r.height) return;
    const nx = clamp(((e.clientX - r.left) / r.width) * 2 - 1, -1, 1);
    const ny = clamp(((e.clientY - r.top) / r.height) * 2 - 1, -1, 1);
    pointer.kick = clamp(pointer.kick + Math.hypot(nx - pointer.tx, ny - pointer.ty) * 0.35, 0, 0.3);
    pointer.tx = nx;
    pointer.ty = ny;
    pointer.hovering = true;
  };
  const onPointerLeave = () => {
    pointer.hovering = false;
  };

  const zone = canvas.parentElement || canvas;
  if (!reducedMotion) {
    zone.addEventListener("pointermove", onPointerMove);
    zone.addEventListener("pointerleave", onPointerLeave);
  }

  // вызывается движком tre каждый кадр перед рендером сцены
  function update(time, delta) {
    const step = clamp(delta || 0.016, 0.001, 0.05);

    // лёгкий дрейф камеры, чтобы фон жил и без мыши
    const driftX = Math.sin(time * 0.4) * 0.06;
    const driftY = Math.cos(time * 0.31) * 0.045;

    if (reducedMotion) {
      camera.position.set(driftX * 0.3, driftY * 0.3, CAM_Z);
      camera.lookAt(0, 0, 0);
      return;
    }

    const ease = 1 - Math.exp(-step * 5);
    pointer.x += (pointer.tx - pointer.x) * ease;
    pointer.y += (pointer.ty - pointer.y) * ease;
    pointer.hover += ((pointer.hovering ? 1 : 0) - pointer.hover) * (1 - Math.exp(-step * 4));
    pointer.kick *= Math.exp(-step * 2.5);

    camera.position.x = pointer.x * 0.62 + driftX;
    camera.position.y = pointer.y * 0.42 + driftY;
    // наведение на панель — лёгкий «наезд» камеры на логотип
    camera.position.z = CAM_Z - pointer.hover * 0.3 + pointer.kick * 0.35;
    camera.lookAt(0, 0, 0);
    // крен и микро-тряска от рывка мыши
    camera.rotation.z += -pointer.x * 0.02 + pointer.kick * 0.06 * Math.sin(time * 40);

    dust.rotation.z = time * 0.02;
    dust.position.y = Math.sin(time * 0.24) * 0.18;
  }

  img.onload = paintLogo;
  img.src = opts.logoSrc || "/static/img/ckkd.png";

  let art;
  try {
    art = createAsciiArt({
      canvas,
      source: { scene, camera, update },
      bg: "#05030d",
      minCellPx: 5,
      lowCost: coarseLowPower(),
      crt: { barrel: 0.035, aberration: 0.0012, vignette: 0.5 },
      onFrame,
    });
  } catch (err) {
    console.warn("ascii-about: WebGL/движок недоступны, фон пропущен", err);
    canvas.dataset.asciiError = (err && err.message) || String(err);
    return () => {};
  }

  // отладочный хэндл (используется тестами и консолью)
  canvas.__treArt = art;
  canvas.__treDebug = { scene, camera, update };

  const frame = art.frame;
  frame.tint.set(VIOLET);
  frame.colorize = 1; // глифы берут собственный фиолетовый логотипа
  frame.persistence = reducedMotion ? 0.15 : 0.55;
  frame.jitter = 0;
  frame.power = 0; // старт с «выключенного» CRT

  const syncCellPx = () => {
    frame.cellPx = clamp(Math.round(art.canvas.width / 130), 5, 11);
  };

  const ro = new ResizeObserver(() => {
    layout();
    syncCellPx();
  });
  ro.observe(canvas);
  layout();
  syncCellPx();

  function onFrame(f, time) {
    if (reducedMotion) {
      f.power = 1;
      f.glitch = 0;
      f.scramble = 0.02;
      f.glow = 1.1;
      return;
    }

    // Плавное «включение» кинескопа при появлении секции.
    const boot = clamp(time / BOOT_SECONDS, 0, 1);
    const bootEase = boot * boot * (3 - 2 * boot);
    f.power = bootEase;

    // Фоновое мерцание + редкие одиночные срывы кадра (datamosh),
    // плюс всплеск от быстрого движения мыши и лёгкое возмущение при наведении.
    const spike = Math.random() < 0.01 ? 0.25 + Math.random() * 0.25 : 0;
    f.glitch =
      (1 - bootEase) * 0.5 +
      0.035 +
      0.03 * (0.5 + 0.5 * Math.sin(time * 0.9) * Math.sin(time * 0.53 + 2)) +
      pointer.kick +
      pointer.hover * 0.03 +
      spike;
    f.jitter = Math.min(1, pointer.kick * 2);
    f.scramble = 0.03 + (1 - bootEase) * 0.3;
    f.glow = 1.15 + 0.2 * Math.sin(time * 0.7) + pointer.hover * 0.6;
  }

  return () => {
    ro.disconnect();
    zone.removeEventListener("pointermove", onPointerMove);
    zone.removeEventListener("pointerleave", onPointerLeave);
    if (logoTexture) logoTexture.dispose();
    logoPlane.geometry.dispose();
    logoPlane.material.dispose();
    dust.geometry.dispose();
    dust.material.dispose();
    art.dispose();
  };
}
