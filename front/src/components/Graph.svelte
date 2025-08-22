<script>
  import { onDestroy, onMount } from "svelte";
  import { socket } from "../stores/socket.js";

  let data = [];
  let unsubscribe;
  let canvas;
  let gl;
  let width = 1024;
  let height = 512;

  let texW;
  let texH;
  let texWF;
  let rowIndex;
  let progBars;
  let bufBars;
  let locPosBars;
  let locValBars;
  let locMinBars;
  let locMaxBars;

  let progWF;
  let bufWF;
  let locPosWF;
  let locTexWF;
  let locRowsWF;

  unsubscribe = socket.subscribe((ws) => {
    if (!ws) return;
    ws.addEventListener("message", handleMessage);
  });

  function handleMessage(e) {
    const msg = JSON.parse(e.data);
    if (msg.type === "graph") {
      data = msg.data;
      drawGraph();
    }
  }

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  onMount(() => {
    setupWebGL();
  });
  // ------------------------------------------------------------------
  // 1)  setupWebGL – create two programs, one texture, one quad buffer
  // ------------------------------------------------------------------
  function setupWebGL() {
    gl = canvas.getContext("webgl");
    if (!gl) {
      alert("WebGL not supported");
      return;
    }

    const compile = (type, src) => {
      const sh = gl.createShader(type);
      gl.shaderSource(sh, src);
      gl.compileShader(sh);
      if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS))
        throw gl.getShaderInfoLog(sh);
      return sh;
    };

    // ------------- bars program -------------
    const vsBars = `
    attribute vec2 a_pos;   // (x,y) in clip space
    attribute float a_val;  // raw value for colouring
    varying float vVal;
    void main() {
      gl_Position = vec4(a_pos, 0.0, 1.0);
      vVal = a_val;
    }`;
    const fsBars = `
    precision mediump float;
    varying float vVal;
    uniform float uMin;
    uniform float uMax;
    vec3 heat(float t){return mix(vec3(0.0,0.0,1.0),vec3(0.0,1.0,0.0),t);}
    void main(){
      float t = clamp((vVal-uMin)/(uMax-uMin),0.0,1.0);
      t = pow(t, 2.5);
      gl_FragColor = vec4(heat(t),1.0);
    }`;

    progBars = gl.createProgram();
    gl.attachShader(progBars, compile(gl.VERTEX_SHADER, vsBars));
    gl.attachShader(progBars, compile(gl.FRAGMENT_SHADER, fsBars));
    gl.linkProgram(progBars);

    bufBars = gl.createBuffer();
    locPosBars = gl.getAttribLocation(progBars, "a_pos");
    locValBars = gl.getAttribLocation(progBars, "a_val");
    locMinBars = gl.getUniformLocation(progBars, "uMin");
    locMaxBars = gl.getUniformLocation(progBars, "uMax");

    // ------------- waterfall program -------------
    const vsWF = `
    attribute vec2 a_pos;
    varying vec2 vUv;
    void main(){
      vUv = a_pos*0.5+0.5;   // (-1..+1) -> (0..1)
      gl_Position = vec4(a_pos,0.0,1.0);
    }`;
    const fsWF = `
    precision mediump float;
    uniform sampler2D u_tex;
    uniform float u_rows;
    varying vec2 vUv;
    void main(){
      float row = (1.0 - vUv.y) * u_rows;
      gl_FragColor = texture2D(u_tex, vec2(vUv.x, row/u_rows));
    }`;

    progWF = gl.createProgram();
    gl.attachShader(progWF, compile(gl.VERTEX_SHADER, vsWF));
    gl.attachShader(progWF, compile(gl.FRAGMENT_SHADER, fsWF));
    gl.linkProgram(progWF);

    bufWF = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, bufWF);
    gl.bufferData(
      gl.ARRAY_BUFFER,
      new Float32Array([-1, -1, 1, -1, -1, 0, 1, 0]),
      gl.STATIC_DRAW,
    );
    locPosWF = gl.getAttribLocation(progWF, "a_pos");
    locTexWF = gl.getUniformLocation(progWF, "u_tex");
    locRowsWF = gl.getUniformLocation(progWF, "u_rows");

    // ------------- waterfall texture -------------
    texW = 100;
    texH = texW /2;
    texWF = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, texWF);
    gl.texImage2D(
      gl.TEXTURE_2D,
      0,
      gl.RGBA,
      texW,
      texH,
      0,
      gl.RGBA,
      gl.UNSIGNED_BYTE,
      null,
    );
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);

    rowIndex = 0; // next row to write
    gl.clearColor(0.0,0.2,0.1,1.0)
  }

  // ------------------------------------------------------------------
  // 2)  drawGraph – bars on top, proper waterfall below
  // ------------------------------------------------------------------
  function drawGraph() {
    if (!gl || !data.length) return;

    gl.clear(gl.COLOR_BUFFER_BIT);

    const N = data.length;
    const minV = Math.min(...data);
    const maxV = Math.max(...data);
    const range = maxV - minV || 1;
    const barW = (2 / N) * 0.9;

    // ------------- 1) Bars (top half y = 0 … 1) -------------
    const barVerts = new Float32Array(N * 6 * 3); // 6 verts * (x,y,val)
    data.forEach((v, i) => {
      const x0 = -1 + i * (2 / N);
      const x1 = x0 + barW;
      const y0 = 0.0; // bottom of bars
      const y1 = (v - minV) / range; // top of bars
      const idx = i * 18;
      barVerts.set([x0, y0, v, x1, y0, v, x0, y1, v], idx);
      barVerts.set([x1, y0, v, x1, y1, v, x0, y1, v], idx + 9);
    });

    gl.useProgram(progBars);
    gl.bindBuffer(gl.ARRAY_BUFFER, bufBars);
    gl.bufferData(gl.ARRAY_BUFFER, barVerts, gl.STREAM_DRAW);

    gl.enableVertexAttribArray(locPosBars);
    gl.vertexAttribPointer(locPosBars, 2, gl.FLOAT, false, 12, 0);
    gl.enableVertexAttribArray(locValBars);
    gl.vertexAttribPointer(locValBars, 1, gl.FLOAT, false, 12, 8);

    gl.uniform1f(locMinBars, minV);
    gl.uniform1f(locMaxBars, maxV);
    gl.drawArrays(gl.TRIANGLES, 0, N * 6);

    // ------------- 2) Waterfall (bottom half y = -1 … 0) -------------
    // scroll: move entire texture one pixel down
    gl.bindTexture(gl.TEXTURE_2D, texWF);
    gl.copyTexSubImage2D(
      gl.TEXTURE_2D,
      0,
      0,
      1, // dst x, y
      0,
      0, // src x, y
      100,
      50,
    );

    // write newest row at top
    const row = new Uint8Array(texW * 4);
    data.forEach((v, i) => {
      const t = Math.max(0, Math.min(1, (v - minV) / range));
      const r = 0;
      const g = Math.round(255 * t);
      
      const b = Math.round(255 * (1 - t));
      row.set([r, g, b, 255], i * 4);
    });
    gl.texSubImage2D(
      gl.TEXTURE_2D,
      0,
      0,
      rowIndex,
      data.length,
      1,
      gl.RGBA,
      gl.UNSIGNED_BYTE,
      row,
    );
    rowIndex = (rowIndex + 1) % texH;

    // draw full-screen quad, but only bottom half is visible
    gl.useProgram(progWF);
    gl.bindBuffer(gl.ARRAY_BUFFER, bufWF);
    gl.enableVertexAttribArray(locPosWF);
    gl.vertexAttribPointer(locPosWF, 2, gl.FLOAT, false, 0, 0);
    gl.bindTexture(gl.TEXTURE_2D, texWF);
    gl.uniform1i(locTexWF, 0);
    gl.uniform1f(locRowsWF, texH);
    gl.drawArrays(gl.TRIANGLE_STRIP, 0, 4);
  }
</script>

<canvas bind:this={canvas} {width} {height}></canvas>

<style>
  canvas {
    border: 1px solid #ccc;
  }
</style>
