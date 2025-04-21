document.addEventListener('DOMContentLoaded', () => {
    const gqlInput = document.getElementById('gql-input');
    const neo4jOutput = document.getElementById('neo4j-output');
    const errorMessage = document.getElementById('error-message');
    const neo4jVersion = document.getElementById('neo4j-version');
    const apocEnabled = document.getElementById('apoc-enabled');
    let debounceTimeout;

    [gqlInput, neo4jVersion, apocEnabled].forEach(input => {
        input.addEventListener('input', () => {
            clearTimeout(debounceTimeout);
            debounceTimeout = setTimeout(convertGQL, 500);
        });
    });


    /**
     * @param {string} text
     * @param {string | boolean} error
     * @returns {Promise<void>}
     */
    async function updateOutput(text, error = false) {
        if (error) {
            errorMessage.textContent = error;
            errorMessage.classList.remove('hidden');
            neo4jOutput.textContent = '';
        } else {
            errorMessage.classList.add('hidden');
            neo4jOutput.textContent = text || 'No conversion result';
            Prism.highlightElement(neo4jOutput);
        }
    }

    /**
     * @returns {Promise<void>}
     */
    async function convertGQL() {
        const gql = gqlInput.value.trim();

        if (!gql) {
            await updateOutput('Neo4j conversion will appear here...');
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
                    apocEnabled: apocEnabled.checked,
                }),
            });

            const data = await response.json();
            await updateOutput(data.neo4j, data.error);
        } catch (error) {
            await updateOutput('Failed to convert: ' + error.message, true);
        }
    }
});
