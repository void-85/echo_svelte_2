<script>
  import { onMount } from 'svelte';
  import { connectSocket } from './stores/socket.js';
  import Graph from './components/Graph.svelte';
  //import Spectrum from './components/Spectrum.svelte';
  //import MapView from './components/MapView.svelte';

  onMount(() => {
    connectSocket('ws://localhost:8080/ws');
  });
</script>

<main>

  <Graph />

</main>
