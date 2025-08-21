<script>
  import { onDestroy, onMount } from "svelte";
  import { socket } from "../stores/socket.js";

  let data = [];
  let unsubscribe;
  let canvas;
  let gl;
  let program;
  let buffer;
  let width = 1000,
    height = 300;
  let valueBuffer;
  let uMinLoc;
  let uMaxLoc;

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

  /*
  function setupWebGL() {
    gl = canvas.getContext("webgl");
    if (!gl) {
      alert("WebGL not supported");
      return;
    }

    // ------------- Vertex shader -------------
    const vertexShaderSource = `
      attribute vec2 a_position;   // x,y coordinates (-1…+1)
      attribute float a_value;     // <-- NEW: the data[i] value you want to colour by

      varying float vValue;        // <-- NEW: will be interpolated across the primitive

      void main() {
        gl_Position = vec4(a_position, 0.0, 1.0);
        vValue = a_value;          // just forward the value
      }
    `;

    // ------------- Fragment shader -------------
    const fragmentShaderSource = `
      precision mediump float;

      varying float vValue;        // receives the interpolated value

      // two helper uniforms so you can rescale the colour without recompiling shaders
      uniform float uMinVal;
      uniform float uMaxVal;

      // a simple colour-map: low → blue, high → red
      vec3 heatmap(float t) {
        return mix(vec3(0.0, 0.0, 1.0),   // blue
                  vec3(1.0, 0.0, 0.0),   // red
                  t);
      }

      void main() {
        float t = clamp((vValue - uMinVal) / (uMaxVal - uMinVal), 0.0, 1.0);
        gl_FragColor = vec4(heatmap(t), 1.0);
      }
    `;

    const vertexShader = compileShader(gl.VERTEX_SHADER, vertexShaderSource);
    const fragmentShader = compileShader(gl.FRAGMENT_SHADER, fragmentShaderSource);

    program = gl.createProgram();
 
    gl.attachShader(program, vertexShader);
    gl.attachShader(program, fragmentShader);
    gl.linkProgram(program);
    gl.useProgram(program);

    buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);

    const a_position = gl.getAttribLocation(program, "a_position");
    gl.enableVertexAttribArray(a_position);
    gl.vertexAttribPointer(a_position, 2, gl.FLOAT, false, 0, 0);

    gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    gl.clearColor(0.0, 0.2, 0.1, 1);
    gl.clear(gl.COLOR_BUFFER_BIT);

  }
*/

  function setupWebGL() {
    gl = canvas.getContext("webgl");
    if (!gl) {
      alert("WebGL not supported");
      return;
    }

    // ------------- Vertex shader -------------
    const vertexShaderSource = `
    attribute vec2 a_position;
    attribute float a_value;
    varying   float vValue;

    void main() {
      gl_Position = vec4(a_position, 0.0, 1.0);
      vValue      = a_value;
    }
  `;

    // ------------- Fragment shader -------------
    const fragmentShaderSource = `
    precision mediump float;

    varying float vValue;
    uniform float uMinVal;
    uniform float uMaxVal;

    vec3 heat(float t) {
      return mix(vec3(1.0, 0.0, 0.0), vec3(0.0, 1.0, 0.0), t);
    }

    void main() {
      float t = clamp((vValue - uMinVal) / (uMaxVal - uMinVal), 0.0, 1.0);
      gl_FragColor = vec4(heat(t), 1.0);
    }
  `;

    const vertexShader = compileShader(gl.VERTEX_SHADER, vertexShaderSource);
    const fragmentShader = compileShader(
      gl.FRAGMENT_SHADER,
      fragmentShaderSource,
    );

    program = gl.createProgram();
    gl.attachShader(program, vertexShader);
    gl.attachShader(program, fragmentShader);
    gl.linkProgram(program);
    gl.useProgram(program);

    // ---------- Position buffer ----------
    buffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    const a_position = gl.getAttribLocation(program, "a_position");
    gl.enableVertexAttribArray(a_position);
    gl.vertexAttribPointer(a_position, 2, gl.FLOAT, false, 0, 0);

    // ---------- Value buffer (colour) ----------
    valueBuffer = gl.createBuffer(); // NEW
    gl.bindBuffer(gl.ARRAY_BUFFER, valueBuffer);
    const a_value = gl.getAttribLocation(program, "a_value");
    gl.enableVertexAttribArray(a_value);
    gl.vertexAttribPointer(a_value, 1, gl.FLOAT, false, 0, 0);

    // ---------- Uniform locations ----------
    uMinLoc = gl.getUniformLocation(program, "uMinVal"); // NEW
    uMaxLoc = gl.getUniformLocation(program, "uMaxVal"); // NEW

    // ---------- Viewport & clear ----------
    gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    gl.clearColor(0.0, 0.2, 0.1, 1);
    gl.clear(gl.COLOR_BUFFER_BIT);
  }

  function compileShader(type, source) {
    const shader = gl.createShader(type);
    gl.shaderSource(shader, source);
    gl.compileShader(shader);
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
      console.error(gl.getShaderInfoLog(shader));
      gl.deleteShader(shader);
    }
    return shader;
  }

  /*function drawGraph() {
    if (!gl || !data.length) return;

    gl.clear(gl.COLOR_BUFFER_BIT);

    // Normalize data into [-1, 1] WebGL clip space
    const xScale = 2 / (data.length - 1);
    const minY = Math.min(...data);
    const maxY = Math.max(...data);
    const yScale = 2 / (maxY - minY || 1);

    const vertices = [];
    data.forEach((d, i) => {
      const x = -1 + i * xScale; // from -1 to +1
      const y = -1 + (d - minY) * yScale; // normalize to -1..+1
      vertices.push(x, y);
    });

    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(vertices), gl.STATIC_DRAW);

    gl.drawArrays(gl.LINE_STRIP, 0, data.length);
  }*/

  function drawGraph() {
    if (!gl || !data.length) return;

    gl.clear(gl.COLOR_BUFFER_BIT);

    // ---- compute scales -------------------------------------------------
    const xScale = 2 / (data.length - 1);
    const minY = Math.min(...data);
    const maxY = Math.max(...data);
    const yRange = maxY - minY || 1; // avoid /0
    const yScale = 2 / yRange;

    // ---- build interleaved positions (x,y) ------------------------------
    const positions = [];
    data.forEach((d, i) => {
      const x = -1 + i * xScale;
      const y = -1 + (d - minY) * yScale;
      positions.push(x, y);
    });

    // ---- upload positions ----------------------------------------------
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(positions), gl.STREAM_DRAW);

    // ---- upload per-point values for colouring --------------------------
    gl.bindBuffer(gl.ARRAY_BUFFER, valueBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data), gl.STREAM_DRAW);

    // ---- set colour-range uniforms --------------------------------------
    gl.uniform1f(uMinLoc, minY);
    gl.uniform1f(uMaxLoc, maxY);

    // ---- draw -----------------------------------------------------------
    gl.drawArrays(gl.LINE_STRIP, 0, data.length);
  }
</script>

<canvas bind:this={canvas} {width} {height}></canvas>

<style>
  canvas {
    border: 1px solid #ccc;
  }
</style>
