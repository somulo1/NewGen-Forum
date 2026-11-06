// Laptops differ in usable height and width, and every size in the app is written in pixels. Rather than letting short screens
// squeeze the layout, the whole interface scales down to fit, the way a slide does. Phones and tablets (900px or narrower) are left alone.
const FULL_W = 1280, FULL_H = 760, SMALLEST = 0.85;

export default defineNuxtPlugin(() => {
  const root = document.documentElement;
  const apply = () => {
    const w = window.innerWidth, h = window.innerHeight;
    const z = w <= 900 ? 1 : Math.round(Math.max(SMALLEST, Math.min(1, w / FULL_W, h / FULL_H)) * 100) / 100;
    (window as any).__uiZ = z;
    root.style.setProperty('--ui-z', String(z));
    root.style.zoom = z === 1 ? '' : String(z);
  };
  apply();
  window.addEventListener('resize', apply);
});
