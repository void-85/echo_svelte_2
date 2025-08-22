<script>
    import { onMount, onDestroy } from "svelte";
    import { socket } from "../stores/socket.js";
    import * as THREE from "three";
    import { log } from "three/tsl";

    // Props for customization
    export let width = 1024;
    export let height = 512;
    export let barCount = 100;

    let canvas;
    let renderer, scene, camera;
    let bars = [];
    let waterfallTexture, waterfallMaterial;
    let values = new Array(barCount).fill(0);
    let animationFrameId;

    let data;
    let unsubscribe;
    let smooth_r, smooth_g, smooth_b

    const texture_y_resolution = 512;

    function handleMessage(e) {
        const msg = JSON.parse(e.data);
        if (msg.type === "graph") {
            data = msg.data;
            updateGraph();
        }
    }

    // Initialize Three.js scene
    function init() {
        if (!canvas) {
            console.error("Canvas not found");
            return;
        }

        // Scene setup
        scene = new THREE.Scene();
        camera = new THREE.OrthographicCamera(
            -width / 2,
            width / 2,
            height / 2,
            -height / 2,
            0.1,
            1000,
        );
        camera.position.set(0, 0, 10); // Set camera position

        renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
        renderer.setSize(width, height);
        renderer.setPixelRatio(window.devicePixelRatio); // Improve rendering quality

        // Spectrum (top half: bars)
        const barWidth = width / barCount;
        for (let i = 0; i < barCount; i++) {
            const geometry = new THREE.PlaneGeometry(barWidth * 0.95, 4);
            const material = new THREE.MeshBasicMaterial({ color: 0x000000 });
            const bar = new THREE.Mesh(geometry, material);
            bar.position.set(-width / 2 + barWidth * (i + 0.5), height / 4, 0);
            scene.add(bar);
            bars.push(bar);
        }

        // Waterfall (bottom half: texture on rectangle)
        // Number of rows in waterfall texture
        const textureWidth = barCount;
        const textureHeight = texture_y_resolution;
        const textureData = new Uint8Array(textureWidth * textureHeight * 4); // RGBA buffer
        waterfallTexture = new THREE.DataTexture(
            textureData,
            textureWidth,
            textureHeight,
            THREE.RGBAFormat,
        );
        //waterfallTexture.minFilter = THREE.NearestFilter;
        //waterfallTexture.magFilter = THREE.NearestFilter;

        waterfallTexture.minFilter = THREE.LinearFilter;
        waterfallTexture.magFilter = THREE.LinearFilter;

        waterfallTexture.needsUpdate = true;

        const geometry = new THREE.PlaneGeometry(width, height * 0.75);
        waterfallMaterial = new THREE.MeshBasicMaterial({
            map: waterfallTexture,
        });
        const waterfallMesh = new THREE.Mesh(geometry, waterfallMaterial);
        waterfallMesh.position.set(0, -height * 0.125, 1);
        waterfallTexture.flipY = true;
        //waterfallMesh.rotateX(3.14/3)
        scene.add(waterfallMesh);

        // Start animation
        animate();
    }

    // Update bar heights and colors, and waterfall texture
    function updateGraph() {
        if (!scene || !camera || !renderer || !waterfallTexture) return;

        // Generate random values for demo (replace with real data if needed)
        //values = values.map(() => Math.random() * height / 4);
        if (data) {
            values = data;

            const data_max = Math.max(...data);
            const data_min = Math.min(...data);

            //console.log(" --- " + data_min + " <-> " + data_max + " ---");

            // Update spectrum bars
            bars.forEach((bar, i) => {
                const value = values[i];
                bar.scale.y = value; // Scale height
                const hue = 1- value / 65 / 2.5; // Map value to hue (0 to 1)
                bar.material.color.setHSL(hue, 1.0, 0.5); // Color based on value
            });

            // Update waterfall texture
            
            const textureData = waterfallTexture.image.data;
            const textureWidth = barCount;
            const textureHeight = texture_y_resolution;

            // Shift existing rows down
            for (let y = textureHeight - 1; y > 0; y--) {
                for (let x = 0; x < textureWidth; x++) {
                    const src = (y - 1) * textureWidth * 4 + x * 4;
                    const dst = (y - 0) * textureWidth * 4 + x * 4;
                    // Ensure we don’t access out-of-bounds
                    if (src >= 0 && dst < textureData.length) {
                        textureData[dst + 0] = textureData[src + 0];
                        textureData[dst + 1] = textureData[src + 1];
                        textureData[dst + 2] = textureData[src + 2];
                        textureData[dst + 3] = textureData[src + 3];
                    }
                }
            }

            // Add new row at top
            for (let x = 0; x < textureWidth; x++) {

                const value = values[x];
                const hue = 1 - value / 65 / 2.5;
                const color = new THREE.Color().setHSL(hue, 1, 0.5);
                const idx = x * 4; // Top row (y=0)



                // Ensure index is within bounds
                if (idx < textureData.length) {
                    textureData[idx + 0] = 255 * color.r;
                    textureData[idx + 1] = 255 * color.g;
                    textureData[idx + 2] = 255 * color.b;
                    textureData[idx + 3] = 255; // Full opacity
                }
            }

            waterfallTexture.needsUpdate = true;
        }
    }

    // Animation loop
    function animate() {
        animationFrameId = requestAnimationFrame(animate);
        updateGraph();
        renderer.render(scene, camera);
    }

    // Lifecycle: Initialize on mount
    onMount(() => {
        init();

        unsubscribe = socket.subscribe((ws) => {
            if (!ws) return;
            ws.addEventListener("message", handleMessage);
        });

        return () => {
            if (animationFrameId) cancelAnimationFrame(animationFrameId);
            if (renderer) {
                renderer.dispose();
                renderer.forceContextLoss();
            }
        };
    });

    // Lifecycle: Clean up on destroy
    onDestroy(() => {
        if (renderer) {
            renderer.dispose();
            renderer.forceContextLoss();
        }
    });
</script>

<canvas bind:this={canvas} style="width: {width}px; height: {height}px;"
></canvas>

<style>
    canvas {
        display: block;
        margin: 0 auto;
    }
</style>
