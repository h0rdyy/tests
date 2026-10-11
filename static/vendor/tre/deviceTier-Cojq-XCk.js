import * as e from "three";
//#region src/acts/acts.ts
function t(e) {
	return e < 0 ? 0 : e > 1 ? 1 : e;
}
function n(e) {
	let n = t(e);
	return n * n * (3 - 2 * n);
}
function r(e) {
	if (e.length === 0) throw Error("defineActs: need at least one act");
	let r = e.reduce((e, t) => e + t.weight, 0), i = [], a = [], o = 0;
	for (let t of e) i.push({
		start: o / r,
		end: (o + t.weight) / r
	}), a.push(o + t.weight / 2), o += t.weight;
	let s = Object.fromEntries(e.map((e, t) => [e.key, t]));
	return {
		acts: e,
		count: e.length,
		total: r,
		index: s,
		bounds: i,
		centers: a,
		actLocal: (e, n) => {
			let r = i[n];
			return t((e - r.start) / (r.end - r.start));
		},
		actIndex: (e) => {
			for (let t = 0; t < i.length; t++) if (e < i[t].end) return t;
			return i.length - 1;
		},
		spanGate: (e, t, r, a = .03) => {
			let o = i[t].start, s = i[r].end, c = n((e - o) / a), l = n((s - e) / a);
			return Math.min(c, l);
		}
	};
}
//#endregion
//#region src/ascii/glyphAtlas.ts
var i = " .·:-=+x*o%#@", a = "ui-monospace, \"SFMono-Regular\", Menlo, monospace";
function o(t, n = 48) {
	let r = typeof t == "string" || t === void 0 ? {
		ramp: t,
		cellW: n
	} : t, i = [...r.ramp ?? " .·:-=+x*o%#@"], o = i.length, s = r.cellW ?? 48, c = s * 2, l = document.createElement("canvas");
	l.width = s * o, l.height = c;
	let u = l.getContext("2d");
	u.fillStyle = "#000", u.fillRect(0, 0, l.width, l.height), u.fillStyle = "#fff", u.font = `${r.weight ?? 400} ${Math.floor(c * .74)}px ${r.font ?? a}`, u.textAlign = "center", u.textBaseline = "middle";
	for (let e = 0; e < o; e++) u.fillText(i[e], e * s + s / 2, c / 2 + c * .02);
	let d = new e.CanvasTexture(l);
	return d.minFilter = e.LinearFilter, d.magFilter = e.LinearFilter, d.generateMipmaps = !1, d.colorSpace = e.NoColorSpace, d.needsUpdate = !0, {
		texture: d,
		count: o
	};
}
//#endregion
//#region src/ascii/FullscreenPass.ts
var s = "\n  varying vec2 vUv;\n  void main() {\n    vUv = uv;\n    gl_Position = vec4(position.xy, 0.0, 1.0);\n  }\n";
function c() {
	let t = new e.BufferGeometry();
	return t.setAttribute("position", new e.BufferAttribute(new Float32Array([
		-1,
		-1,
		0,
		3,
		-1,
		0,
		-1,
		3,
		0
	]), 3)), t.setAttribute("uv", new e.BufferAttribute(new Float32Array([
		0,
		0,
		2,
		0,
		0,
		2
	]), 2)), t;
}
var l = new e.Camera(), u = class {
	material;
	uniforms;
	scene;
	geometry;
	constructor(t, n, r) {
		this.uniforms = n, this.material = new e.ShaderMaterial({
			vertexShader: s,
			fragmentShader: t,
			uniforms: n,
			defines: r,
			depthTest: !1,
			depthWrite: !1
		}), this.geometry = c();
		let i = new e.Mesh(this.geometry, this.material);
		i.frustumCulled = !1, this.scene = new e.Scene(), this.scene.add(i);
	}
	render(e, t) {
		e.setRenderTarget(t), e.render(this.scene, l);
	}
	dispose() {
		this.material.dispose(), this.geometry.dispose();
	}
}, d = "\n  uniform sampler2D uScene;\n  uniform sampler2D uGlyphs;\n  uniform float uCount;\n  uniform float uCellPx;\n  uniform vec3 uTint;\n  uniform vec3 uBg;\n  uniform float uScan;\n  uniform float uScramble;\n  uniform float uColorize;\n  uniform float uTime;\n  uniform vec2 uResolution;\n  varying vec2 vUv;\n\n  float hash(vec2 p) {\n    return fract(sin(dot(p, vec2(41.3, 289.1))) * 43758.5453);\n  }\n\n  void main() {\n    vec2 cells = uResolution / vec2(uCellPx, uCellPx * 2.0);\n\n    vec2 cellId = floor(vUv * cells);\n    vec2 cellCenter = (cellId + 0.5) / cells;\n\n    // overload / boot scramble: nudge the sample point so cells smear\n    if (uScramble > 0.0) {\n      vec2 j = (vec2(hash(cellId + 1.7), hash(cellId + 7.3)) - 0.5) * uScramble / cells;\n      cellCenter += j;\n    }\n\n    vec3 scene = texture2D(uScene, cellCenter).rgb;\n    float lum = clamp(max(scene.r, max(scene.g, scene.b)), 0.0, 1.0);\n    lum = pow(lum, 0.85);\n\n    float gi = floor(lum * (uCount - 0.001));\n    // ...and swap a fraction of lit glyphs for garbage characters\n    if (uScramble > 0.0 && lum > 0.04) {\n      float r = hash(cellId * 1.3 + floor(uTime * 12.0));\n      if (r < uScramble * 0.7) gi = floor(hash(cellId + floor(uTime * 9.0)) * uCount);\n    }\n    vec2 local = fract(vUv * cells);\n    vec2 gUv = vec2((gi + local.x) / uCount, local.y);\n    float mask = texture2D(uGlyphs, gUv).r;\n\n    // ink: phosphor tint, optionally re-inked with the source's own chroma\n    // (uColorize 1 = full-colour ASCII; brightness still comes from lum below)\n    vec3 ink = uTint;\n    if (uColorize > 0.001) {\n      vec3 chroma = scene / max(max(scene.r, max(scene.g, scene.b)), 1e-4);\n      ink = mix(uTint, chroma, uColorize);\n    }\n\n    // phosphor: blend background -> ink by glyph coverage, brighter where lit\n    vec3 col = mix(uBg, ink, mask * (0.35 + 0.65 * lum));\n\n    // per-cell horizontal seam (the coarse \"text rows\" shading)\n    float scan = 1.0 - uScan * 0.5 * (0.5 + 0.5 * sin(cellId.y * 3.14159 + 1.57));\n    col *= scan;\n\n    gl_FragColor = vec4(col, 1.0);\n  }\n", f = "\n  uniform sampler2D uSrc;\n  uniform vec2 uScale;\n  uniform vec2 uOffset;\n  uniform vec3 uFill;\n  varying vec2 vUv;\n\n  void main() {\n    vec2 uv = (vUv - 0.5) * uScale + 0.5 + uOffset;\n    if (uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0) {\n      gl_FragColor = vec4(uFill, 1.0);\n    } else {\n      gl_FragColor = vec4(texture2D(uSrc, uv).rgb, 1.0);\n    }\n  }\n", p = "\n  uniform sampler2D uCurr;\n  uniform sampler2D uPrev;\n  uniform float uDecay;\n  varying vec2 vUv;\n\n  void main() {\n    vec3 c = texture2D(uCurr, vUv).rgb;\n    vec3 p = texture2D(uPrev, vUv).rgb * uDecay;\n    gl_FragColor = vec4(max(c, p), 1.0);\n  }\n", m = "\n  uniform sampler2D uSrc;\n  uniform vec2 uResolution;\n  uniform vec2 uTexel;\n  uniform float uBarrel;\n  uniform float uAberr;\n  uniform float uGlow;\n  uniform float uScan;\n  uniform float uVig;\n  uniform float uTime;\n  uniform float uPower;\n  uniform float uGlitch;\n  uniform float uJitter;\n  varying vec2 vUv;\n\n  float hash(vec2 p) {\n    return fract(sin(dot(p, vec2(41.3, 289.1))) * 43758.5453);\n  }\n\n  void main() {\n    vec2 uv = vUv;\n    float gl = uGlitch;\n\n    // datamosh: shove whole rows sideways at random; degauss jitters the frame\n    float row = floor(uv.y * 26.0);\n    float tearGate = step(0.62, hash(vec2(row * 1.7, floor(uTime * 11.0))));\n    float blockShift = (hash(vec2(row, floor(uTime * 17.0))) - 0.5) * gl * 0.14 * tearGate;\n    vec2 jit = (vec2(hash(vec2(floor(uTime * 31.0), 1.0)), hash(vec2(3.0, floor(uTime * 27.0)))) - 0.5)\n               * (uJitter * 0.05);\n    vec2 uvj = uv + vec2(blockShift, 0.0) + jit;\n\n    vec2 c = uvj - 0.5;\n    float r2 = dot(c, c);\n\n    // barrel curvature — bulge the screen outward\n    vec2 duv = uvj + c * r2 * uBarrel;\n\n    // chromatic aberration, stronger toward the edges and during a glitch\n    float a = (uAberr + gl * 0.02) * (0.4 + r2);\n    vec3 base;\n    base.r = texture2D(uSrc, duv + c * a).r;\n    base.g = texture2D(uSrc, duv).g;\n    base.b = texture2D(uSrc, duv - c * a).b;\n\n    // phosphor glow — ring of taps, bright areas bleed into a halo\n    vec3 glow = vec3(0.0);\n    vec2 px = uTexel * 2.2;\n    #ifdef GLOW_LOW\n      for (int i = 0; i < 4; i++) {\n        float ang = float(i) * 1.5707963;\n        vec2 o = vec2(cos(ang), sin(ang));\n        glow += texture2D(uSrc, duv + o * px * 1.7).rgb;\n      }\n      glow /= 4.0;\n    #else\n      for (int i = 0; i < 8; i++) {\n        float ang = float(i) * 0.785398;\n        vec2 o = vec2(cos(ang), sin(ang));\n        glow += texture2D(uSrc, duv + o * px).rgb;\n        glow += texture2D(uSrc, duv + o * px * 2.0).rgb * 0.5;\n      }\n      glow /= 12.0;\n    #endif\n\n    vec3 col = base + glow * uGlow;\n\n    // signal-loss flicker during a glitch\n    col *= 1.0 - gl * 0.28 * hash(vec2(row, floor(uTime * 20.0)));\n\n    // rounded screen edge (outside the curved frame = black)\n    float edge = (duv.x < 0.0 || duv.x > 1.0 || duv.y < 0.0 || duv.y > 1.0) ? 0.0 : 1.0;\n    col *= edge;\n\n    // scanlines (anchored to the un-jittered uv so they don't swim)\n    col *= 1.0 - uScan * 0.4 * (0.5 + 0.5 * sin(uv.y * uResolution.y * 3.14159));\n\n    // vignette\n    col *= clamp(1.0 - uVig * r2 * 1.35, 0.0, 1.0);\n\n    // CRT power envelope: collapse to a horizontal line, then a dot, then black\n    float powerH = smoothstep(0.0, 0.5, uPower);\n    float powerV = mix(0.015, 1.0, smoothstep(0.5, 1.0, uPower));\n    vec2 d2 = abs(uv - 0.5) * 2.0;\n    float mask = smoothstep(powerH + 0.012, powerH - 0.012, d2.x)\n               * smoothstep(powerV + 0.012, powerV - 0.012, d2.y);\n    // bright phosphor scanline while collapsed/opening\n    float lineGlow = (1.0 - smoothstep(0.0, 0.05, d2.y)) * (1.0 - powerV) * powerH;\n    col = col * mask + vec3(0.55, 1.0, 0.7) * lineGlow * 0.7;\n\n    gl_FragColor = vec4(col, 1.0);\n    #include <colorspace_fragment>\n  }\n", h = class {
	frame;
	atlasTexture;
	ascii;
	fit;
	persist;
	crtPass;
	gridRT;
	asciiRT;
	persistRTs;
	minCellPx;
	size = new e.Vector2(0, 0);
	probe = new e.Vector2();
	prevIndex = 0;
	time = 0;
	constructor(t = {}) {
		let { ramp: n, font: r, weight: i, bg: a = "#04060b", minCellPx: s = 4, lowCost: c = !1, crt: l, scan: h = 0 } = t;
		this.minCellPx = s;
		let g = o({
			ramp: n,
			font: r,
			weight: i
		});
		this.atlasTexture = g.texture;
		let _ = new e.Color(a);
		this.ascii = new u(d, {
			uScene: { value: null },
			uGlyphs: { value: g.texture },
			uCount: { value: g.count },
			uCellPx: { value: 8 },
			uTint: { value: new e.Color("#9bf6b0") },
			uBg: { value: _ },
			uScan: { value: h },
			uScramble: { value: 0 },
			uColorize: { value: 0 },
			uTime: { value: 0 },
			uResolution: { value: new e.Vector2(1, 1) }
		}), this.fit = new u(f, {
			uSrc: { value: null },
			uScale: { value: new e.Vector2(1, 1) },
			uOffset: { value: new e.Vector2(0, 0) },
			uFill: { value: _ }
		}), this.persist = new u(p, {
			uCurr: { value: null },
			uPrev: { value: null },
			uDecay: { value: 0 }
		}), this.crtPass = new u(m, {
			uSrc: { value: null },
			uResolution: { value: new e.Vector2(1, 1) },
			uTexel: { value: new e.Vector2(1, 1) },
			uBarrel: { value: l?.barrel ?? .12 },
			uAberr: { value: l?.aberration ?? .0016 },
			uGlow: { value: .9 },
			uScan: { value: l?.scan ?? .35 },
			uVig: { value: l?.vignette ?? .6 },
			uTime: { value: 0 },
			uPower: { value: 1 },
			uGlitch: { value: 0 },
			uJitter: { value: 0 }
		}, c ? { GLOW_LOW: "1" } : void 0), this.gridRT = new e.WebGLRenderTarget(4, 4, {
			samples: c ? 0 : 4,
			depthBuffer: !0,
			minFilter: e.LinearFilter,
			magFilter: e.LinearFilter,
			generateMipmaps: !1
		});
		let v = {
			depthBuffer: !1,
			minFilter: e.LinearFilter,
			magFilter: e.LinearFilter,
			generateMipmaps: !1
		};
		this.asciiRT = new e.WebGLRenderTarget(4, 4, v), this.persistRTs = [new e.WebGLRenderTarget(4, 4, v), new e.WebGLRenderTarget(4, 4, v)], this.frame = {
			cellPx: 8,
			tint: new e.Color("#9bf6b0"),
			colorize: 0,
			scramble: 0,
			power: 1,
			glitch: 0,
			jitter: 0,
			persistence: 0,
			glow: .9
		};
	}
	setRamp(e) {
		let t = o(typeof e == "string" ? { ramp: e } : e);
		this.atlasTexture.dispose(), this.atlasTexture = t.texture, this.ascii.uniforms.uGlyphs.value = t.texture, this.ascii.uniforms.uCount.value = t.count;
	}
	updateSize(e) {
		if (e.getDrawingBufferSize(this.probe), this.probe.equals(this.size)) return;
		this.size.copy(this.probe);
		let t = Math.max(2, this.size.x), n = Math.max(2, this.size.y);
		this.gridRT.setSize(Math.max(2, Math.ceil(t / this.minCellPx)), Math.max(2, Math.ceil(n / (this.minCellPx * 2)))), this.asciiRT.setSize(t, n), this.persistRTs[0].setSize(t, n), this.persistRTs[1].setSize(t, n), this.ascii.uniforms.uResolution.value.set(t, n), this.crtPass.uniforms.uResolution.value.set(t, n), this.crtPass.uniforms.uTexel.value.set(1 / t, 1 / n);
	}
	renderScene(e, t, n, r, i) {
		this.updateSize(e), e.setRenderTarget(this.gridRT), e.render(t, n), this.compose(e, r, i);
	}
	renderTexture(e, t, n, r, i = "cover") {
		this.updateSize(e);
		let a = this.fit.uniforms.uScale.value;
		if (a.set(1, 1), i !== "fill") {
			let e = t.image, n = (e?.videoWidth || e?.width || 1) / (e?.videoHeight || e?.height || 1), r = this.size.x / Math.max(1, this.size.y) / n;
			i === "cover" ? r < 1 ? a.set(r, 1) : a.set(1, 1 / r) : r < 1 ? a.set(1, 1 / r) : a.set(r, 1);
		}
		this.fit.uniforms.uSrc.value = t, this.fit.render(e, this.gridRT), this.compose(e, n, r);
	}
	compose(e, t, n) {
		let { ascii: r, persist: i, crtPass: a, gridRT: o, asciiRT: s, persistRTs: c, frame: l } = this;
		this.time = t, r.uniforms.uCellPx.value = l.cellPx, r.uniforms.uTint.value.copy(l.tint), r.uniforms.uScramble.value = l.scramble, r.uniforms.uColorize.value = l.colorize, r.uniforms.uTime.value = this.time, a.uniforms.uTime.value = this.time, a.uniforms.uPower.value = l.power, a.uniforms.uGlitch.value = l.glitch, a.uniforms.uJitter.value = l.jitter, a.uniforms.uGlow.value = l.glow, r.uniforms.uScene.value = o.texture, r.render(e, s);
		let u = s.texture;
		if (l.persistence > .01) {
			let t = c[this.prevIndex], r = c[1 - this.prevIndex];
			i.uniforms.uCurr.value = s.texture, i.uniforms.uPrev.value = t.texture, i.uniforms.uDecay.value = l.persistence ** (n * 60), i.render(e, r), this.prevIndex = 1 - this.prevIndex, u = r.texture;
		}
		a.uniforms.uSrc.value = u, a.render(e, null);
	}
	dispose() {
		this.atlasTexture.dispose(), this.ascii.dispose(), this.fit.dispose(), this.persist.dispose(), this.crtPass.dispose(), this.gridRT.dispose(), this.asciiRT.dispose(), this.persistRTs[0].dispose(), this.persistRTs[1].dispose();
	}
}, g = {
	classic: " .·:-=+x*o%#@",
	minimal: " .:+*#",
	dense: " .`'\":;Il!i~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$",
	blocks: " ░▒▓█",
	braille: " ⠁⠃⠇⠏⠟⠿⡿⣿",
	katakana: " ･ｰｨｸｺﾆﾈﾎﾒﾋﾜｷﾑ",
	binary: " .01",
	lines: " ─│┼╬█"
};
//#endregion
//#region src/art/sources.ts
function _(e) {
	return typeof e == "object" && !!e && "scene" in e && "camera" in e;
}
function v(t) {
	if (_(t)) return {
		kind: "scene",
		scene: t,
		dynamic: !1,
		owned: !1
	};
	if (t instanceof e.Texture) return {
		kind: "texture",
		texture: t,
		dynamic: !1,
		owned: !1
	};
	if (typeof HTMLVideoElement < "u" && t instanceof HTMLVideoElement) {
		let n = new e.VideoTexture(t);
		return n.colorSpace = e.SRGBColorSpace, {
			kind: "texture",
			texture: n,
			dynamic: !1,
			owned: !0
		};
	}
	if (typeof HTMLImageElement < "u" && t instanceof HTMLImageElement) {
		let n = new e.Texture(t);
		return n.colorSpace = e.SRGBColorSpace, n.generateMipmaps = !0, n.minFilter = e.LinearMipmapLinearFilter, n.magFilter = e.LinearFilter, t.complete && t.naturalWidth > 0 ? n.needsUpdate = !0 : t.addEventListener("load", () => n.needsUpdate = !0, { once: !0 }), {
			kind: "texture",
			texture: n,
			dynamic: !1,
			owned: !0
		};
	}
	if (typeof ImageBitmap < "u" && t instanceof ImageBitmap) {
		let n = new e.Texture(t);
		return n.colorSpace = e.SRGBColorSpace, n.flipY = !1, n.needsUpdate = !0, {
			kind: "texture",
			texture: n,
			dynamic: !1,
			owned: !0
		};
	}
	let n = new e.CanvasTexture(t);
	return n.colorSpace = e.SRGBColorSpace, n.generateMipmaps = !1, n.minFilter = e.LinearFilter, {
		kind: "texture",
		texture: n,
		dynamic: !0,
		owned: !0
	};
}
async function y(e = { facingMode: "user" }) {
	let t = await navigator.mediaDevices.getUserMedia({
		video: e,
		audio: !1
	}), n = document.createElement("video");
	return n.srcObject = t, n.muted = !0, n.playsInline = !0, await n.play(), {
		video: n,
		stream: t,
		stop: () => {
			n.pause(), n.srcObject = null;
			for (let e of t.getTracks()) e.stop();
		}
	};
}
//#endregion
//#region src/art/createAsciiArt.ts
function b(t = {}) {
	let { canvas: n = document.createElement("canvas"), fit: r = "cover", fps: i = 60, dpr: a = Math.min(typeof devicePixelRatio == "number" ? devicePixelRatio : 1, 1.5), autoStart: o = !0, autoPause: s = !0, onFrame: c, source: l, ...u } = t, d = new e.WebGLRenderer({
		canvas: n,
		antialias: !1,
		alpha: !1,
		powerPreference: "high-performance",
		stencil: !1
	});
	d.setPixelRatio(a);
	let f = new h(u), p = l ? v(l) : null, m = o, g = !0, _ = 0, y = !1, b = 0, x = 0, S = 0, C = !1, w = 1 / Math.max(1, i), T = !0, E = () => {
		T = !1;
		let e = n.clientWidth || n.width || 2, t = n.clientHeight || n.height || 2;
		d.setSize(e, t, !1);
	}, D = typeof ResizeObserver < "u" ? new ResizeObserver(() => T = !0) : null;
	D?.observe(n);
	let O = (e) => {
		T && E(), !(n.clientWidth === 0 && n.width <= 2) && (c?.(f.frame, S, e), p && (p.kind === "scene" && p.scene ? (p.scene.update?.(S, e), f.renderScene(d, p.scene.scene, p.scene.camera, S, e)) : p.texture && (p.dynamic && (p.texture.needsUpdate = !0), f.renderTexture(d, p.texture, S, e, r))));
	}, k = (e) => {
		_ = requestAnimationFrame(k);
		let t = Math.min((e - b) / 1e3, .1);
		if (b = e, x += t, x < w) return;
		let n = x;
		x = 0, S += n, O(n);
	}, A = () => {
		let e = !C && m && (!s || g && !document.hidden);
		e && !y ? (y = !0, b = performance.now(), _ = requestAnimationFrame(k)) : !e && y && (y = !1, cancelAnimationFrame(_));
	}, j = s && typeof IntersectionObserver < "u" ? new IntersectionObserver((e) => {
		g = e[e.length - 1].isIntersecting, A();
	}) : null;
	j?.observe(n);
	let M = () => A();
	s && document.addEventListener("visibilitychange", M);
	let N = () => {
		p?.owned && p.texture?.dispose(), p = null;
	};
	return A(), {
		canvas: n,
		renderer: d,
		pipeline: f,
		frame: f.frame,
		get running() {
			return y;
		},
		setSource(e) {
			N(), p = e ? v(e) : null;
		},
		setRamp(e) {
			f.setRamp(e);
		},
		start() {
			m = !0, A();
		},
		stop() {
			m = !1, A();
		},
		renderOnce() {
			O(w);
		},
		snapshot(e = "image/png", t) {
			return O(w), n.toDataURL(e, t);
		},
		dispose() {
			C = !0, A(), D?.disconnect(), j?.disconnect(), s && document.removeEventListener("visibilitychange", M), N(), f.dispose(), d.dispose();
		}
	};
}
//#endregion
//#region src/audio/audioReactor.ts
var x = {
	level: 0,
	bass: 0,
	mid: 0,
	treble: 0,
	beat: 0,
	active: !1
}, S = 2048, C = .18, w = class {
	ctx = null;
	analyser = null;
	freq = null;
	source = null;
	stream = null;
	elementSources = /* @__PURE__ */ new WeakMap();
	raf = 0;
	lastT = 0;
	peak = .12;
	bassEMA = 0;
	lastBeatT = 0;
	kind = null;
	ensureContext() {
		return this.ctx || (this.ctx = new AudioContext(), this.analyser = this.ctx.createAnalyser(), this.analyser.fftSize = S, this.analyser.smoothingTimeConstant = .65, this.freq = new Uint8Array(this.analyser.frequencyBinCount)), this.ctx.state === "suspended" && this.ctx.resume(), this.ctx;
	}
	async connectMic() {
		let e = this.ensureContext(), t = await navigator.mediaDevices.getUserMedia({ audio: {
			echoCancellation: !1,
			noiseSuppression: !1,
			autoGainControl: !1
		} });
		this.attach(e.createMediaStreamSource(t), "mic", t, !1);
	}
	async connectDisplayAudio() {
		let e = this.ensureContext(), t = await navigator.mediaDevices.getDisplayMedia({
			video: !0,
			audio: !0
		});
		for (let e of t.getVideoTracks()) e.stop();
		if (t.getAudioTracks().length === 0) {
			for (let e of t.getTracks()) e.stop();
			throw Error("no audio track shared — tick \"share audio\" in the picker");
		}
		this.attach(e.createMediaStreamSource(t), "display", t, !1);
	}
	connectElement(e) {
		let t = this.ensureContext(), n = this.elementSources.get(e);
		n || (n = t.createMediaElementSource(e), this.elementSources.set(e, n)), this.attach(n, "element", null, !0);
	}
	disconnect() {
		if (this.source && this.analyser) try {
			this.source.disconnect();
		} catch {}
		if (this.stream) {
			for (let e of this.stream.getTracks()) e.stop();
			this.stream = null;
		}
		this.source = null, this.kind = null, x.active = !1, cancelAnimationFrame(this.raf), this.raf = 0, x.level = x.bass = x.mid = x.treble = x.beat = 0;
	}
	dispose() {
		this.disconnect(), this.ctx?.close(), this.ctx = null, this.analyser = null;
	}
	attach(e, t, n, r) {
		this.disconnect();
		let i = this.ctx, a = this.analyser;
		if (e.connect(a), r) a.connect(i.destination);
		else try {
			a.disconnect();
		} catch {}
		this.source = e, this.stream = n, this.kind = t, x.active = !0, this.lastT = performance.now(), this.raf ||= requestAnimationFrame(this.tick);
	}
	band(e, t, n) {
		let r = this.ctx.sampleRate / S, i = Math.max(0, Math.floor(t / r)), a = Math.min(e.length - 1, Math.ceil(n / r)), o = 0;
		for (let t = i; t <= a; t++) o += e[t];
		return o / ((a - i + 1) * 255);
	}
	tick = (e) => {
		if (!x.active || !this.analyser || !this.freq) {
			this.raf = 0;
			return;
		}
		this.raf = requestAnimationFrame(this.tick);
		let t = Math.min(.1, (e - this.lastT) / 1e3);
		this.lastT = e, this.analyser.getByteFrequencyData(this.freq);
		let n = this.band(this.freq, 20, 250), r = this.band(this.freq, 250, 2e3), i = this.band(this.freq, 2e3, 9e3), a = Math.max(n, r * .9, i * .8);
		this.peak = Math.max(this.peak * (1 - t * .08), a, .12);
		let o = 1 / this.peak, s = (e, t) => e + (t - e) * (t > e ? .5 : .12);
		x.bass = s(x.bass, Math.min(1, n * o)), x.mid = s(x.mid, Math.min(1, r * o)), x.treble = s(x.treble, Math.min(1, i * o)), x.level = s(x.level, Math.min(1, a * o)), this.bassEMA += (x.bass - this.bassEMA) * .04;
		let c = e / 1e3;
		x.bass > this.bassEMA * 1.35 + .03 && c - this.lastBeatT > C ? (x.beat = 1, this.lastBeatT = c) : (x.beat *= Math.exp(-t * 6), x.beat < .001 && (x.beat = 0));
	};
}, T = new w();
//#endregion
//#region src/perf/deviceTier.ts
function E() {
	if (typeof navigator > "u") return !1;
	let e = navigator, t = e.hardwareConcurrency ?? 4, n = e.deviceMemory ?? 4;
	return /Android|iPhone|iPad|iPod|Mobile/i.test(e.userAgent) || t <= 4 || n <= 4;
}
function D(e) {
	return e <= 480 ? 1.5 : e <= 768 ? 1.25 : 1;
}
//#endregion
export { i as _, T as a, r as b, _ as c, h as d, d as f, u as g, p as h, x as i, v as l, f as m, E as n, b as o, m as p, w as r, y as s, D as t, g as u, o as v, n as x, t as y };

//# sourceMappingURL=deviceTier-Cojq-XCk.js.map