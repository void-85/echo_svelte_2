<script>
  import { onDestroy } from 'svelte';
  import { socket } from '../stores/socket.js';
  import * as d3 from 'd3';

  let data = [];
  let unsubscribe;

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

  function drawGraph() {
    if (!data.length) return;

    console.log("[ drawing graph... ]")

    const svg = d3.select("svg");
    svg.selectAll('*').remove();
    const width = 600, height = 400;
    const x = d3.scaleLinear().domain([0, data.length - 1]).range([0, width]);
    const y = d3.scaleLinear().domain([d3.min(data), d3.max(data)]).range([height, 0]);
    const line = d3.line().x((d, i) => x(i)).y(d => y(d));
    svg.append('path')
      .datum(data)
      .attr('fill', 'none')
      .attr('stroke', 'steelblue')
      .attr('stroke-width', 2)
      .attr('d', line);
  }
</script>

<svg width={600} height={400}></svg>

<style>
  svg { border: 1px solid #ccc; }
</style>
