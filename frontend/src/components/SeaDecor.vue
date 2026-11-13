<script setup>
// Trang trí "đáy biển" cố định ở hai góc dưới màn hình: san hô + rong biển lắc nhẹ, bong bóng khí nổi lên chậm.
// Nằm SAU nội dung (z-index thấp, không nhận chuột) nên không che thẻ sản phẩm; chân trang (nền đậm) phủ lên khi cuộn tới.
// Bật "giảm chuyển động" → san hô đứng yên, không có bong bóng.
const bubbles = {
  left: [
    { x: 38, s: 7, t: 13, d: 0 }, { x: 70, s: 5, t: 16, d: 4.5 }, { x: 104, s: 9, t: 14, d: 8 },
    { x: 58, s: 4, t: 18, d: 11 }, { x: 128, s: 6, t: 15, d: 2.5 },
  ],
  right: [
    { x: 46, s: 6, t: 15, d: 1.5 }, { x: 92, s: 8, t: 13, d: 6 }, { x: 140, s: 5, t: 17, d: 9.5 },
    { x: 70, s: 4, t: 19, d: 13 }, { x: 118, s: 7, t: 14, d: 3.5 },
  ],
}
</script>

<template>
  <div class="sea" aria-hidden="true">
    <!-- Cụm trái -->
    <div class="cluster left">
      <svg viewBox="0 0 220 150" preserveAspectRatio="xMinYMax meet">
        <ellipse cx="96" cy="150" rx="110" ry="14" class="rock" />
        <path class="weed sway-a" d="M24 150 C14 124 34 108 24 82 C16 60 30 44 26 22" />
        <path class="weed sway-b thin" d="M40 150 C34 132 48 118 42 100 C36 84 46 72 44 58" />
        <g class="coral sway-c">
          <path d="M84 150 L84 112" /><path d="M84 124 Q72 104 64 88" /><path d="M84 114 Q96 94 104 80" />
          <path d="M64 88 L58 70" /><path d="M64 88 L74 72" /><path d="M104 80 L100 62" /><path d="M104 80 L116 68" />
        </g>
        <g class="fan sway-a">
          <path d="M156 150 C150 118 130 104 132 84 C134 66 154 58 166 60 C182 62 194 76 190 96 C186 116 166 124 160 150 Z" />
          <path class="vein" d="M158 146 L150 90 M160 140 L166 72 M162 132 L182 86" />
        </g>
        <circle cx="200" cy="140" r="7" class="pebble" /><circle cx="186" cy="144" r="4" class="pebble" />
      </svg>
      <span v-for="(b, i) in bubbles.left" :key="'l' + i" class="bubble"
        :style="{ left: b.x + 'px', width: b.s + 'px', height: b.s + 'px', '--t': b.t + 's', '--d': b.d + 's' }" />
    </div>
    <!-- Cụm phải -->
    <div class="cluster right">
      <svg viewBox="0 0 220 150" preserveAspectRatio="xMaxYMax meet">
        <ellipse cx="124" cy="150" rx="110" ry="14" class="rock" />
        <g class="coral alt sway-b">
          <path d="M60 150 L60 116" /><path d="M60 126 Q48 110 42 96" /><path d="M60 118 Q72 100 78 88" />
          <path d="M42 96 L38 80" /><path d="M78 88 L74 72" /><path d="M78 88 L90 78" />
        </g>
        <path class="weed sway-c" d="M118 150 C108 126 128 110 118 86 C110 66 124 50 120 30" />
        <path class="weed sway-a thin" d="M134 150 C128 134 142 120 136 104 C130 90 140 80 138 66" />
        <g class="coral sway-a">
          <path d="M178 150 L178 120" /><path d="M178 128 Q168 112 162 100" /><path d="M178 122 Q190 108 196 96" />
          <path d="M162 100 L158 86" /><path d="M196 96 L200 82" />
        </g>
        <circle cx="20" cy="142" r="6" class="pebble" /><circle cx="34" cy="146" r="4" class="pebble" />
      </svg>
      <span v-for="(b, i) in bubbles.right" :key="'r' + i" class="bubble"
        :style="{ right: b.x + 'px', width: b.s + 'px', height: b.s + 'px', '--t': b.t + 's', '--d': b.d + 's' }" />
    </div>
  </div>
</template>

<style scoped>
.sea { position: fixed; left: 0; right: 0; bottom: 0; height: 0; z-index: 0; pointer-events: none; }
.cluster { position: absolute; bottom: 0; width: 220px; height: 150px; opacity: .72; }
.cluster.left { left: 0; }
.cluster.right { right: 0; }
svg { width: 100%; height: 100%; overflow: visible; display: block; }

.rock { fill: #d5e0ee; }
.pebble { fill: #c3d2e5; }
.weed { fill: none; stroke: #7fbfb0; stroke-width: 7; stroke-linecap: round; }
.weed.thin { stroke-width: 5; stroke: #99cfc2; }
.coral path { fill: none; stroke: #f2a68f; stroke-width: 8; stroke-linecap: round; stroke-linejoin: round; }
.coral.alt path { stroke: #f4b98a; }
.fan > path:first-child { fill: #f7caa0; }
.fan .vein { fill: none; stroke: #eeae7c; stroke-width: 2.5; stroke-linecap: round; }

/* Lắc nhẹ như có dòng nước, gốc ở đáy */
.sway-a, .sway-b, .sway-c { transform-box: fill-box; transform-origin: 50% 100%; animation: sway 7s ease-in-out infinite alternate; }
.sway-b { animation-duration: 9s; animation-delay: -3s; }
.sway-c { animation-duration: 11s; animation-delay: -6s; }
@keyframes sway { from { transform: rotate(-3deg); } to { transform: rotate(3deg); } }

/* Bong bóng: trong, viền xanh nhạt, có đốm sáng; nổi chậm, lượn nhẹ rồi tan */
.bubble { position: absolute; bottom: 60px; border-radius: 50%; opacity: 0;
  background: radial-gradient(circle at 32% 30%, rgba(255,255,255,.95) 0 18%, rgba(255,255,255,.15) 40%, rgba(47,124,240,.08) 70%);
  border: 1px solid rgba(11,94,215,.28);
  animation: bubble var(--t, 14s) ease-in var(--d, 0s) infinite; }
@keyframes bubble {
  0% { opacity: 0; transform: translate(0, 0); }
  10% { opacity: .85; }
  30% { transform: translate(6px, -12vh); }
  55% { transform: translate(-5px, -24vh); }
  80% { opacity: .5; transform: translate(4px, -34vh); }
  100% { opacity: 0; transform: translate(0, -42vh); }
}

@media (max-width: 767px) {
  .cluster { width: 130px; height: 90px; opacity: .6; }
  .bubble { bottom: 36px; }
  .bubble:nth-of-type(n+4) { display: none; }
}
@media (prefers-reduced-motion: reduce) {
  .bubble { display: none; }
  .sway-a, .sway-b, .sway-c { animation: none; }
}
</style>
