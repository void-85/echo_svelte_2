<script>
  import { onDestroy, onMount } from 'svelte';
  import { socket } from '../stores/socket.js';

  let data = [];
  let unsubscribe;
  let canvas;
  let gl;
  let program;
  let buffer;
  let width = 1000, height = 500;

  unsubscribe = socket.subscribe(ws => {
    if (!ws) return;
    ws.addEventListener('message', handleMessage);
  });

  function handleMessage(e) {
    const msg = JSON.parse(e.data);
    if (msg.type === 'graph') {
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

  function setupWebGL() {
    gl = canvas.getContext("webgl");
    if (!gl) {
      alert("WebGL not supported");
      return;
    }

    // Simple vertex + fragment shaders
    const vertexShaderSource = `
      attribute vec2 a_position;
      void main() {
        gl_Position = vec4(a_position, 0, 1);
      }
    `;
    const fragmentShaderSource = `
      precision mediump float;
      void main() {
        gl_FragColor = vec4(0.27, 0.51, 0.71, 1.0); // steelblue
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
    gl.clearColor(1, 1, 1, 1);
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

  function drawGraph() {
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
  }
</script>

<canvas bind:this={canvas} width={width} height={height}></canvas>

<style>
  canvas {
    border: 1px solid #ccc;
  }
</style>
