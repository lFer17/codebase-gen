    document.addEventListener('DOMContentLoaded', () => {
        const form = document.getElementById('generator-form');
        const generateBtn = document.getElementById('generate-btn');
        const resultSection = document.getElementById('result-section');
        const console = document.getElementById('console');
        const downloadSection = document.getElementById('download-section');
        const downloadLink = document.getElementById('download-link');

        let websocket = null;

        form.addEventListener('submit', (e) => {
            e.preventDefault();
            startGeneration();
        });

        function startGeneration() {
            // Show results section and clear previous output
            resultSection.classList.remove('hidden');
            console.innerHTML = '';
            downloadSection.classList.add('hidden');

            // Disable submit button
            generateBtn.disabled = true;
            generateBtn.innerText = 'Generating...';

            // Get form values
            const prompt = document.getElementById('prompt').value;
            const language = document.getElementById('language').value;
            const template = document.getElementById('template').value;
            const basePackage = document.getElementById('base-package').value;
            const workerCount = document.getElementById('worker-count').value;
            const model = document.getElementById('model').value;

            // Connect to WebSocket
            websocket = new WebSocket(`ws://${window.location.host}/api/generate`);

            websocket.onopen = () => {
                // Send request
                websocket.send(JSON.stringify({
                    prompt,
                    language,
                    template,
                    basePackage,
                    workerCount: parseInt(workerCount),
                    model,
                    projectName: document.getElementById('project-name').value || `${language}-project`
                }));

                log('info', 'Connected to server. Starting code generation...');
            };

            websocket.onmessage = (event) => {
                const data = JSON.parse(event.data);

                switch(data.type) {
                    case 'start':
                        log('info', data.message);
                        break;
                    case 'file':
                        log('info', `Writing file: ${data.file}`);
                        break;
                    case 'error':
                        log('error', `Error: ${data.error}`);
                        generateBtn.disabled = false;
                        generateBtn.innerText = 'Generate Code';
                        break;
                    case 'complete':
                        log('success', data.message);
                        downloadLink.href = data.zipUrl;
                        downloadSection.classList.remove('hidden');
                        generateBtn.disabled = false;
                        generateBtn.innerText = 'Generate Code';
                        break;
                }
            };

            websocket.onerror = (error) => {
                log('error', `WebSocket error: ${error}`);
                generateBtn.disabled = false;
                generateBtn.innerText = 'Generate Code';
            };

            websocket.onclose = () => {
                log('info', 'Connection closed');
            };
        }

        function log(type, message) {
            const p = document.createElement('p');
            p.classList.add(type);
            p.innerText = message;
            console.appendChild(p);
            console.scrollTop = console.scrollHeight;
        }
    });