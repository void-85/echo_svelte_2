<script>
  import { onMount } from 'svelte';
  import { connectSocket } from './stores/socket.js';
  import Graph from './components/Graph.svelte';
    import Waterfall from './components/Waterfall.svelte';
  //import Spectrum from './components/Spectrum.svelte';
  //import MapView from './components/MapView.svelte';

  onMount(() => {
    connectSocket('ws://localhost:8080/ws');
  });
</script>

<main>

  <h3>TEST PAGE _2</h3>
  <!-- <Graph/> -->

  <br>
  <br>
  <br>
  <br>

  NEW WATERFALL GRAPH:
  <Waterfall/>

</main>
