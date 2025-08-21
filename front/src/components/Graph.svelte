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

  let bufX;
  let bufHeight;
  let bufValue;
  let locX;
  let locHeight;
  let locValue;

  let locPos
  let locVal

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

    // Vertex shader: 2-D position + the value we colour by
    const vs = `
    attribute vec2 a_pos;   // (x,y) bottom-left / top-right corners
    attribute float a_val;  // raw data value
    varying   float vVal;

    void main() {
      gl_Position = vec4(a_pos, 0.0, 1.0);
      vVal = a_val;
    }
  `;

    // Fragment shader: colour from value
    const fs = `
    precision mediump float;
    varying float vVal;
    uniform float uMinVal;
    uniform float uMaxVal;

    vec3 heat(float t) {
      return mix(vec3(0.2, 0.2, 0.5), vec3(0.0, 1.0, 0.0), t);
    }

    void main() {
      float t = clamp((vVal - uMinVal) / (uMaxVal - uMinVal), 0.0, 1.0);
      //float t = smoothstep(uMinVal, uMaxVal, vVal);

      t = pow(t, 4.0);

      gl_FragColor = vec4(heat(t), 1.0);
    }
  `;

    const compile = (type, src) => {
      const s = gl.createShader(type);
      gl.shaderSource(s, src);
      gl.compileShader(s);
      if (!gl.getShaderParameter(s, gl.COMPILE_STATUS))
        throw gl.getShaderInfoLog(s);
      return s;
    };

    const vsObj = compile(gl.VERTEX_SHADER, vs);
    const fsObj = compile(gl.FRAGMENT_SHADER, fs);

    program = gl.createProgram();
    gl.attachShader(program, vsObj);
    gl.attachShader(program, fsObj);
    gl.linkProgram(program);
    gl.useProgram(program);

    // One interleaved buffer:  (x,y,value) * 6 vertices per bar
    buffer = gl.createBuffer();

    // Attribute & uniform locations
    locPos = gl.getAttribLocation(program, "a_pos");
    locVal = gl.getAttribLocation(program, "a_val");
    uMinLoc = gl.getUniformLocation(program, "uMinVal");
    uMaxLoc = gl.getUniformLocation(program, "uMaxVal");

    gl.viewport(0, 0, gl.canvas.width, gl.canvas.height);
    gl.clearColor(0.1, 0.1, 0.0, 1.0);
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

    const N = data.length;
    const minVal = Math.min(...data);
    const maxVal = Math.max(...data);
    const range = maxVal - minVal || 1;

    // pixel width of one bar in clip space (constant bar width)
    const barW = (2 / N) * 0.9; // 0.9 => thin gap between bars

    // Build one interleaved array:  (x,y,value) for each vertex
    // Two triangles per bar = 6 vertices
    const verts = new Float32Array(N * 6 * 3); // 6 verts * 3 floats (x,y,val)

    data.forEach((v, i) => {
      const x0 = -1 + i * (2 / N); // left
      const x1 = x0 + barW; // right
      const y0 = -1; // bottom (value 0)
      const y1 = -1 + ((v - minVal) / range) * 2; // top (scaled to NDC)

      const idx = i * 18; // 6 verts * 3 floats
      // Triangle 1
      verts.set([x0, y0, v, x1, y0, v, x0, y1, v], idx);
      // Triangle 2
      verts.set([x1, y0, v, x1, y1, v, x0, y1, v], idx + 9);
    });

    // Upload
    gl.bindBuffer(gl.ARRAY_BUFFER, buffer);
    gl.bufferData(gl.ARRAY_BUFFER, verts, gl.STREAM_DRAW);

    // Tell WebGL how to read the buffer:
    //   stride = 3 * 4 = 12 bytes, offset = 0 for pos, 8 for value
    gl.enableVertexAttribArray(locPos);
    gl.vertexAttribPointer(locPos, 2, gl.FLOAT, false, 12, 0);

    gl.enableVertexAttribArray(locVal);
    gl.vertexAttribPointer(locVal, 1, gl.FLOAT, false, 12, 8);

    // Colour scale
    gl.uniform1f(uMinLoc, minVal);
    gl.uniform1f(uMaxLoc, maxVal);

    // Draw
    gl.drawArrays(gl.TRIANGLES, 0, N * 6);
  }
</script>

<canvas bind:this={canvas} {width} {height}></canvas>

<style>
  canvas {
    border: 1px solid #ccc;
  }
</style>
