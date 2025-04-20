document.addEventListener('DOMContentLoaded', () => {
    const gqlInput = document.getElementById('gql-input');
    const neo4jOutput = document.getElementById('neo4j-output');
    const errorMessage = document.getElementById('error-message');
    const neo4jVersion = document.getElementById('neo4j-version');
    const apocEnabled = document.getElementById('apoc-enabled');
    let debounceTimeout;

    // Load Neo4j versions
    async function loadNeo4jVersions() {
        try {
            const response = await fetch('/api/neo4j-versions');
            const versions = await response.json();

            // Clear loading option
            neo4jVersion.innerHTML = '';

            // Add versions to select (already filtered and sorted by backend)
            versions.forEach(version => {
                const option = document.createElement('option');
                option.value = version;
                option.textContent = version;
                neo4jVersion.appendChild(option);
            });

            // Select first version (which will be latest due to server-side sorting)
            if (versions.length > 0) {
                neo4jVersion.value = versions[0];
                convertGQL(); // Trigger initial conversion
            }
        } catch (error) {
            console.error('Failed to load Neo4j versions:', error);
            const option = document.createElement('option');
            option.value = '';
            option.textContent = 'Failed to load versions';
            neo4jVersion.innerHTML = '';
            neo4jVersion.appendChild(option);
        }
    }

    // Add event listeners for all inputs
    [gqlInput, neo4jVersion, apocEnabled].forEach(input => {
        input.addEventListener('input', () => {
            clearTimeout(debounceTimeout);
            debounceTimeout = setTimeout(convertGQL, 500);
        });
    });

    async function updateOutput(text, isError = false) {
        if (isError) {
            errorMessage.textContent = text;
            errorMessage.classList.remove('hidden');
            neo4jOutput.textContent = '';
        } else {
            errorMessage.classList.add('hidden');
            neo4jOutput.textContent = text || 'No conversion result';
            // Trigger Prism highlighting
            Prism.highlightElement(neo4jOutput);
        }
    }

    async function convertGQL() {
        const gql = gqlInput.value.trim();

        if (!gql) {
            updateOutput('Neo4j conversion will appear here...');
            return;
        }

        try {
            const response = await fetch('/convert', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    gql,
                    neo4jVersion: neo4jVersion.value,
                    apocEnabled: apocEnabled.checked
                }),
            });

            const data = await response.json();
            updateOutput(data.neo4j, Boolean(data.error));
        } catch (error) {
            updateOutput('Failed to convert: ' + error.message, true);
        }
    }

    // Load versions when the page loads
    loadNeo4jVersions();
});
